package space

import (
	"context"
	"database/sql"
	"errors"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/google/uuid"
)

func (db *spaceRepo) GetSpaceByID(ctx context.Context, id uuid.UUID) (model.Space, error) {
	res := model.Space{}

	err := db.db.QueryRowContext(ctx, "select id, name, created, creator, personal from shared_spaces.shared_spaces where id = $1", id).
		Scan(&res.ID, &res.Name, &res.Created, &res.Creator, &res.Personal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Space{}, api_errors.ErrSpaceNotExists
		}
	}

	return res, err
}
