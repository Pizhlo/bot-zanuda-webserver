package note

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"webserver/internal/model"

	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"github.com/sirupsen/logrus"
)

type noteRepo struct {
	db        *sql.DB
	currentTx *sql.Tx
}

func New(addr string) (*noteRepo, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, fmt.Errorf("connect open a db driver: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to a db: %w", err)
	}
	return &noteRepo{db, nil}, nil
}

func (db *noteRepo) Close() {
	if err := db.db.Close(); err != nil {
		logrus.Errorf("error on closing note repo: %v", err)
	}
}

func (db *noteRepo) tx(ctx context.Context) (*sql.Tx, error) {
	if db.currentTx != nil {
		return db.currentTx, nil
	}

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, err
	}

	db.currentTx = tx

	return tx, nil
}

func (db *noteRepo) commit() error {
	tx := db.currentTx
	db.currentTx = nil
	return tx.Commit()
}

func (db *noteRepo) rollback() error {
	tx := db.currentTx
	db.currentTx = nil
	return tx.Rollback()
}

// ошибка о том, что пользователя не существует в БД
var ErrUnknownUser = errors.New("unknown user")

// ошибка о том, что пространства не существует
var ErrSpaceNotExists = errors.New("space not exists")

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
