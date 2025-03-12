package space

import (
	"context"
	"errors"
	"fmt"
	"webserver/internal/model"
)

func (db *spaceRepo) GetSpaceByID(ctx context.Context, id int) (model.Space, error) {
	res := model.Space{}

	err := db.db.QueryRowContext(ctx, "select id, name, created, creator, personal from shared_spaces.shared_spaces where id = $1", id).
		Scan(&res.ID, &res.Name, &res.Created, &res.Creator, &res.Personal)
	return res, err
}

var (
	// ошибка о том, что у пользователя нет заметок
	ErrNoNotesFoundByUserID = errors.New("user does not have any notes")
)

// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает в полном виде.
func (db *spaceRepo) GetAllbySpaceIDFull(ctx context.Context, spaceID int64) ([]model.Note, error) {
	res := []model.Note{}

	rows, err := db.db.QueryContext(ctx, `select  notes.notes.id as note_id, text as note_text, notes.notes.created as note_created, 
	last_edit as note_last_edit, shared_spaces.shared_spaces.id as space_id,  shared_spaces.shared_spaces.name as space_name, 
	shared_spaces.shared_spaces.personal, shared_spaces.shared_spaces.creator,shared_spaces.shared_spaces.created as space_created, 
	users.users.tg_id,  users.users.username,  users.users.space_id as users_personal_space, users.timezones.timezone 
	from shared_spaces.shared_spaces
left join notes.notes on shared_spaces.shared_spaces.id = notes.notes.space_id
join users.users on users.users.id = notes.notes.user_id
join users.timezones on users.timezones.user_id = notes.notes.user_id
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
			return nil, fmt.Errorf("error scanning note: %+v", err)
		}

		note.User.PersonalSpace = *note.Space

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, ErrNoNotesFoundByUserID
	}

	return res, nil
}

// GetAllbySpaceID возвращает все заметки пользователя из его личного пространства. Информацию о пользователе возвращает кратко (только userID)
func (db *spaceRepo) GetAllBySpaceID(ctx context.Context, spaceID int64) ([]model.GetNote, error) {
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
			return nil, fmt.Errorf("error scanning note: %+v", err)
		}

		res = append(res, note)
	}

	if len(res) == 0 {
		return nil, ErrNoNotesFoundByUserID
	}

	return res, nil
}
