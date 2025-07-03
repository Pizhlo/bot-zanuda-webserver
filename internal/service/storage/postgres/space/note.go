package space

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"webserver/internal/model"
	"webserver/internal/model/elastic"
	"webserver/internal/model/rabbit"

	api_errors "webserver/internal/errors"
	"webserver/internal/service/storage/elasticsearch"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// GetAllNotesBySpaceIDFull возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
func (db *Repo) GetAllNotesBySpaceIDFull(ctx context.Context, spaceID uuid.UUID) ([]model.Note, error) {
	res := []model.Note{}

	rows, err := db.db.QueryContext(ctx, `select  notes.notes.id as note_id, text as note_text, notes.notes.created as note_created, 
	updated as note_updated, shared_spaces.shared_spaces.id as space_id,  shared_spaces.shared_spaces.name as space_name, 
	shared_spaces.shared_spaces.personal, shared_spaces.shared_spaces.creator,shared_spaces.shared_spaces.created as space_created, 
	users.users.tg_id,  users.users.username,  users.users.space_id as users_personal_space, users.timezones.timezone 
	from shared_spaces.shared_spaces
left join notes.notes on shared_spaces.shared_spaces.id = notes.notes.space_id
left join users.users on users.users.id = notes.notes.user_id
left join users.timezones on users.timezones.user_id = notes.notes.user_id
where shared_spaces.shared_spaces.id = $1;`, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error getting all notes by user id: %+v", err)
	}

	for rows.Next() {
		note := model.Note{
			Space: &model.Space{},
			User: &model.User{
				PersonalSpace: &model.Space{},
			},
		}
		err := rows.Scan(&note.ID, &note.Text, &note.Created, &note.Updated,
			&note.Space.ID, &note.Space.Name, &note.Space.Personal,
			&note.Space.Creator, &note.Space.Created, &note.User.TgID,
			&note.User.UsernameSQL, &note.User.PersonalSpace.ID, &note.User.Timezone)
		if err != nil {
			// "sql: Scan error on column index 1, name \"note_text\": converting NULL to string is unsupported"
			if strings.Contains(err.Error(), "converting NULL to string is unsupported") {
				return nil, api_errors.ErrNoNotesFoundBySpaceID
			}

			return nil, fmt.Errorf("error scanning note: %+v", err)
		}

		note.User.PersonalSpace = note.Space
		note.User.Username = note.User.UsernameSQL.String

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, api_errors.ErrSpaceNotExists
	}

	return res, nil
}

// GetAllNotesBySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (db *Repo) GetAllNotesBySpaceID(ctx context.Context, spaceID uuid.UUID) ([]model.GetNote, error) {
	res := []model.GetNote{}

	rows, err := db.db.QueryContext(ctx, `select  notes.notes.id as note_id, text as note_text, notes.notes.created as note_created, updated as note_updated, shared_spaces.shared_spaces.id as space_id,  users.users.tg_id from shared_spaces.shared_spaces
left join notes.notes on shared_spaces.shared_spaces.id = notes.notes.space_id
left join users.users on users.users.id = notes.notes.user_id
left join users.timezones on users.timezones.user_id = notes.notes.user_id
where shared_spaces.shared_spaces.id = $1;`, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error getting all notes by user id: %+v", err)
	}

	for rows.Next() {
		note := model.GetNote{}

		err := rows.Scan(&note.ID, &note.Text, &note.Created, &note.Updated,
			&note.SpaceID, &note.UserID,
		)
		if err != nil {
			// "sql: Scan error on column index 1, name \"note_text\": converting NULL to string is unsupported"
			if strings.Contains(err.Error(), "converting NULL") {
				return nil, api_errors.ErrNoNotesFoundBySpaceID
			}

			return nil, fmt.Errorf("error scanning note: %+v", err)
		}

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, api_errors.ErrSpaceNotExists
	}

	return res, nil
}

func (db *Repo) UpdateNote(ctx context.Context, update rabbit.UpdateNoteRequest) error {
	tx, err := db.tx(ctx)
	if err != nil {
		return err
	}

	var id uuid.UUID
	err = tx.QueryRowContext(ctx, `update notes.notes set text = $1, updated = now()
	where id = $2 and user_id = (select id from users.users where tg_id = $3) returning id`,
		update.Text, update.NoteID, update.UserID).Scan(&id)
	if err != nil {
		return fmt.Errorf("error while updating note: %+v", err)
	}

	data := elastic.Data{
		Index: elastic.NoteIndex,
		Model: &elastic.Note{
			ID:      id,
			TgID:    update.UserID,
			Text:    update.Text,
			SpaceID: update.SpaceID,
		},
	}

	err = db.elasticClient.UpdateNote(ctx, data)
	if err != nil {
		rollBackErr := tx.Rollback()
		if rollBackErr != nil {
			logrus.Errorf("error rollback tx: %+v", rollBackErr)
		}

		return fmt.Errorf("error while updating note in elastic: %+v", err)
	}

	return tx.Commit()
}

// GetNoteByID возвращает заметку по айди, либо ошибку о том, что такой заметки не существует
func (db *Repo) GetNoteByID(ctx context.Context, noteID uuid.UUID) (model.GetNote, error) {
	var note model.GetNote

	row := db.db.QueryRowContext(ctx, `select notes.notes.id, tg_id, text, notes.notes.space_id, created, updated, type
	 from notes.notes 
left join users.users on users.users.id = notes.notes.user_id
where notes.notes.id = $1;`, noteID)

	err := row.Scan(&note.ID, &note.UserID, &note.Text, &note.SpaceID, &note.Created, &note.Updated, &note.Type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.GetNote{}, api_errors.ErrNoteNotFound
		}

		return model.GetNote{}, err
	}

	if note.ID == uuid.Nil {
		return model.GetNote{}, api_errors.ErrNoteNotFound
	}

	return note, nil
}

// GetNotesTypes возвращает все типы заметок в пространстве и их количество (3 текстовых, 2 фото, и т.п.)
func (db *Repo) GetNotesTypes(ctx context.Context, spaceID uuid.UUID) ([]model.NoteTypeResponse, error) {
	res := []model.NoteTypeResponse{}

	rows, err := db.db.QueryContext(ctx, "select count(*), type from notes.notes group by type, notes.space_id having space_id = $1;", spaceID)
	if err != nil {
		return nil, fmt.Errorf("error getting note types: %+v", err)
	}

	for rows.Next() {
		noteType := model.NoteTypeResponse{}

		err := rows.Scan(&noteType.Count, &noteType.Type)
		if err != nil {
			return nil, fmt.Errorf("error scanning result of note types query: %+v", err)
		}

		res = append(res, noteType)
	}

	if len(res) == 0 {
		return nil, api_errors.ErrNoNotesFoundBySpaceID
	}

	return res, nil
}

// GetNotesByType возвращает все заметки указанного типа из пространства
func (db *Repo) GetNotesByType(ctx context.Context, spaceID uuid.UUID, noteType model.NoteType) ([]model.GetNote, error) {
	res := []model.GetNote{}

	rows, err := db.db.QueryContext(ctx, `select notes.notes.id, users.users.tg_id, text, created, updated, file 
from notes.notes
join users.users on users.users.id = notes.notes.user_id
where notes.notes.space_id = $1 and type = $2;`, spaceID, noteType)
	if err != nil {
		return nil, fmt.Errorf("error getting note types: %+v", err)
	}

	for rows.Next() {
		note := model.GetNote{
			SpaceID: spaceID,
			Type:    noteType,
		}

		err := rows.Scan(&note.ID, &note.UserID, &note.Text, &note.Created, &note.Updated, &note.File)
		if err != nil {
			return nil, fmt.Errorf("error scanning result of note types query: %+v", err)
		}

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, api_errors.ErrNoNotesFoundByType
	}

	return res, nil
}

func (db *Repo) SearchNoteByText(ctx context.Context, req model.SearchNoteByTextRequest) ([]model.GetNote, error) {
	search := elastic.Data{
		Index: elastic.NoteIndex,
		Model: &elastic.Note{
			SpaceID: req.SpaceID,
			Text:    req.Text,
			Type:    req.Type,
		},
	}

	ids, err := db.elasticClient.SearchByText(ctx, search)
	if err != nil {
		if errors.Is(err, elasticsearch.ErrRecordsNotFound) {
			return nil, api_errors.ErrNoNotesFoundByText
		}

		return nil, err
	}

	var notes []model.GetNote

	q, args, err := sqlx.In(`select notes.notes.id, users.users.tg_id, text, created, updated, type, file from notes.notes
	join users.users on users.users.id = notes.notes.user_id
	where notes.notes.id IN(?);`, ids)
	if err != nil {
		return nil, fmt.Errorf("error while creating query while searching notes: %+v", err)
	}

	q = sqlx.Rebind(sqlx.DOLLAR, q)
	rows, err := db.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("error while searching notes by text: %w", err)
	}

	for rows.Next() {
		note := model.GetNote{
			SpaceID: req.SpaceID,
		}

		err := rows.Scan(&note.ID, &note.UserID, &note.Text, &note.Created, &note.Updated, &note.Type, &note.File)
		if err != nil {
			return nil, fmt.Errorf("error while scanning note (search by text): %w", err)
		}

		notes = append(notes, note)
	}

	if len(notes) == 0 {
		return nil, api_errors.ErrNoNotesFoundByText
	}

	return notes, nil
}
