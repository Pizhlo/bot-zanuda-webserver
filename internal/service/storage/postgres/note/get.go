package note

import (
	"context"
	"webserver/internal/model"
)

func (db *noteRepo) GetAllbyUserID(ctx context.Context, userID int64) ([]model.GetNoteResponse, error) {
	return nil, nil
}
