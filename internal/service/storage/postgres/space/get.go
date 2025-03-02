package space

import (
	"context"
	"webserver/internal/model"
)

func (db *spaceRepo) GetSpaceByID(ctx context.Context, id int) (model.Space, error) {
	res := model.Space{}

	err := db.db.QueryRowContext(ctx, "select id, name, created, creator, personal from shared_spaces.shared_spaces where id = $1", id).
		Scan(&res.ID, &res.Name, &res.Created, &res.Creator, &res.Personal)
	return res, err
}
