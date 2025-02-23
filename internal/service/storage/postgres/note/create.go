package note

import (
	"context"
	"errors"
	"fmt"
	"webserver/internal/model"
	"webserver/internal/model/elastic"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	// ошибка о том, что пользователя не существует в БД
	ErrUnknownUser = errors.New("unknown user")
	// ошибка о том, что пространства не существует
	ErrSpaceNotExists = errors.New("space not exists")
	// ошибка о том, что пользователь не может добавить запись в это пространство: оно личное и не принадлежит ему
	ErrSpaceNotBelongsUser = errors.New("space not belongs to user")
)

func (db *noteRepo) Create(ctx context.Context, note model.CreateNoteRequest) error {
	tx, err := db.tx(ctx)
	if err != nil {
		return fmt.Errorf("error while creating transaction: %w", err)
	}

	// id новой заметки после создания; нужен для сохранения в elastic
	var noteID uuid.UUID
	err = tx.QueryRowContext(ctx, `insert into notes.notes (user_id, text, space_id, created) values((select id from users.users where tg_id=$1), $2, $3, to_timestamp($4)) returning ID`,
		note.UserID, note.Text, note.SpaceID, note.Created).Scan(&noteID)
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
			ID:   noteID,
			Text: note.Text,
			TgID: note.UserID,
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
