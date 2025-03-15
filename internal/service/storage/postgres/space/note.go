package space

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"webserver/internal/model"
	"webserver/internal/model/elastic"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	// ошибка о том, что пользователя не существует в БД
	ErrUnknownUser = errors.New("unknown user")
	// ошибка о том, что пользователь не может добавить запись в это пространство: оно личное и не принадлежит ему
	ErrSpaceNotBelongsUser = errors.New("space not belongs to user")
	// ошибка о том, что в пространстве нет заметок
	ErrNoNotesFoundBySpaceID = errors.New("space does not have any notes")
	// ошибка о том, что пространство не существует
	ErrSpaceNotExists = errors.New("space does not exist")
)

func (db *spaceRepo) CreateNote(ctx context.Context, note model.CreateNoteRequest) error {
	tx, err := db.tx(ctx)
	if err != nil {
		return fmt.Errorf("error while creating transaction: %w", err)
	}

	// id новой заметки после создания; нужен для сохранения в elastic
	var noteID uuid.UUID
	err = tx.QueryRowContext(ctx, `insert into notes.notes (user_id, text, space_id, created) values((select id from users.users where tg_id=$1), $2, $3, now()) returning ID`,
		note.UserID, note.Text, note.SpaceID).Scan(&noteID)
	if err != nil {
		db.currentTx = nil

		switch t := err.(type) {
		case *pq.Error:
			if t.Code == "23502" && t.Column == "user_id" { // null value in column \"user_id\" of relation \"notes\" violates not-null constraint
				return ErrUnknownUser
			}

			if t.Code == "23503" && t.Constraint == "notes_space_id" {
				return ErrSpaceNotExists
			}

			if t.Code == "P0001" && t.Where == "PL/pgSQL function check_personal_space() line 9 at RAISE" {
				return ErrSpaceNotBelongsUser
			}
		}

		return err
	}

	// создаем структуру для сохранения в elastic
	elasticData := elastic.Data{
		Index: elastic.NoteIndex,
		Model: &elastic.Note{
			ID:      noteID,
			Text:    note.Text,
			TgID:    note.UserID,
			SpaceID: note.SpaceID,
		}}

	// сохраняем в elastic
	err = db.elasticClient.Save(ctx, elasticData)
	if err != nil {
		// отменяем транзакцию в случае ошибки (для консистентности данных)
		_ = tx.Rollback()
		return fmt.Errorf("error saving note to Elastic: %+v", err)
	}

	return db.commit()
}

// GetAllNotesBySpaceIDFull возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
func (db *spaceRepo) GetAllNotesBySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error) {
	res := []model.Note{}

	rows, err := db.db.QueryContext(ctx, `select  notes.notes.id as note_id, text as note_text, notes.notes.created as note_created, 
	last_edit as note_last_edit, shared_spaces.shared_spaces.id as space_id,  shared_spaces.shared_spaces.name as space_name, 
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
			User:  &model.User{},
		}
		err := rows.Scan(&note.ID, &note.Text, &note.Created, &note.LastEdit,
			&note.Space.ID, &note.Space.Name, &note.Space.Personal,
			&note.Space.Creator, &note.Space.Created, &note.User.TgID,
			&note.User.Username, &note.User.PersonalSpace.ID, &note.User.Timezone)
		if err != nil {
			// "sql: Scan error on column index 1, name \"note_text\": converting NULL to string is unsupported"
			if strings.Contains(err.Error(), "converting NULL") {
				return nil, ErrNoNotesFoundBySpaceID
			}

			return nil, fmt.Errorf("error scanning note: %+v", err)
		}

		note.User.PersonalSpace = *note.Space

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, ErrSpaceNotExists
	}

	return res, nil
}

// GetAllNotesBySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (db *spaceRepo) GetAllNotesBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error) {
	res := []model.GetNote{}

	rows, err := db.db.QueryContext(ctx, `select  notes.notes.id as note_id, text as note_text, notes.notes.created as note_created, last_edit as note_last_edit, shared_spaces.shared_spaces.id as space_id,  users.users.tg_id from shared_spaces.shared_spaces
left join notes.notes on shared_spaces.shared_spaces.id = notes.notes.space_id
left join users.users on users.users.id = notes.notes.user_id
left join users.timezones on users.timezones.user_id = notes.notes.user_id
where shared_spaces.shared_spaces.id = $1;`, spaceID)
	if err != nil {
		return nil, fmt.Errorf("error getting all notes by user id: %+v", err)
	}

	for rows.Next() {
		note := model.GetNote{}

		err := rows.Scan(&note.ID, &note.Text, &note.Created, &note.LastEdit,
			&note.SpaceID, &note.UserID,
		)
		if err != nil {
			// "sql: Scan error on column index 1, name \"note_text\": converting NULL to string is unsupported"
			if strings.Contains(err.Error(), "converting NULL") {
				return nil, ErrNoNotesFoundBySpaceID
			}

			return nil, fmt.Errorf("error scanning note: %+v", err)
		}

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, ErrSpaceNotExists
	}

	return res, nil
}

func (db *spaceRepo) UpdateNote(ctx context.Context, update model.UpdateNote) error {
	tx, err := db.tx(ctx)
	if err != nil {
		return err
	}

	var id uuid.UUID
	err = tx.QueryRowContext(ctx, `update notes.notes set text = $1, last_edit = now()
	where id = $2 and user_id = (select id from users.users where tg_id = $3) returning id`,
		update.Text, update.ID, update.UserID).Scan(&id)
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
