package note

import (
	"context"
	"errors"
	"fmt"
	"webserver/internal/model"
)

var (
	// ошибка о том, что у пользователя нет заметок
	ErrNoNotesFoundByUserID = errors.New("user does not have any notes")
)

func (db *noteRepo) GetAllbyUserID(ctx context.Context, userID int64) ([]model.Note, error) {
	res := []model.Note{}

	rows, err := db.db.QueryContext(ctx, `select notes.notes.id, text, notes.notes.created, last_edit, 
	shared_spaces.shared_spaces.id, shared_spaces.shared_spaces.name, shared_spaces.shared_spaces.personal, 
	shared_spaces.shared_spaces.creator, shared_spaces.shared_spaces.created,  users.users.id,  users.users.tg_id,  
	users.users.username,  users.users.space_id, users.timezones.timezone  from notes.notes
join users.users on users.users.id = notes.notes.user_id
join users.timezones on users.timezones.user_id = notes.notes.user_id
join shared_spaces.shared_spaces on shared_spaces.shared_spaces.id = notes.notes.space_id
where notes.notes.user_id = (select id from users.users where tg_id = $1)
and notes.notes.space_id = (select id from shared_spaces.shared_spaces where creator = (select id from users.users where tg_id = $1) and personal = true);`, userID)
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
			&note.Space.Creator, &note.Space.Created, &note.User.ID, &note.User.TgID,
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
