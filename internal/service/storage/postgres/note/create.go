package note

import (
	"context"
	"errors"
	"fmt"
	"webserver/internal/model"

	"github.com/lib/pq"
)

var (
	// ошибка о том, что пользователя не существует в БД
	ErrUnknownUser = errors.New("unknown user")
	// ошибка о том, что пространства не существует
	ErrSpaceNotExists = errors.New("space not exists")
)

func (db *noteRepo) Create(ctx context.Context, note model.CreateNoteRequest) error {
	tx, err := db.tx(ctx)
	if err != nil {
		return fmt.Errorf("error while creating transaction: %w", err)
	}

	_, err = tx.ExecContext(ctx, `insert into notes.notes (user_id, text, space_id, created) values((select id from users.users where tg_id=$1), $2, $3, to_timestamp($4)) returning ID`, note.UserID, note.Text, note.SpaceID, note.Created)
	if err != nil {
		switch t := err.(type) {
		case *pq.Error:
			if t.Code == "23502" && t.Column == "user_id" { // null value in column \"user_id\" of relation \"notes\" violates not-null constraint
				return ErrUnknownUser
			}

			if t.Code == "23503" && t.Constraint == "notes_space_id" {
				return ErrSpaceNotExists
			}
		}

		// err := db.rollback()
		// if err != nil {
		// 	logrus.Errorf("error rollback tx: %+v", err)
		// }

		return err
	}

	return db.commit()
}
