package user

import (
	"context"
	"database/sql"
	"errors"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
)

func (db *userRepo) GetUser(ctx context.Context, tgID int64) (model.User, error) {
	res := model.User{
		PersonalSpace: &model.Space{},
	}

	err := db.db.QueryRowContext(ctx, "select id, tg_id, username, space_id from users.users where tg_id = $1", tgID).
		Scan(&res.ID, &res.TgID, &res.UsernameSQL, &res.PersonalSpace.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, api_errors.ErrUnknownUser
		}

		return model.User{}, err
	}

	res.Username = res.UsernameSQL.String

	return res, err
}

func (db *userRepo) CheckUser(ctx context.Context, tgID int64) (bool, error) {
	var exists bool
	err := db.db.QueryRowContext(ctx, "select exists(select 1 from users.users where tg_id = $1)", tgID).
		Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
