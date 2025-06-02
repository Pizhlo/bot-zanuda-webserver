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
		Scan(&res.ID, &res.ID, &res.TgID, &res.Username, &res.PersonalSpace.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, api_errors.ErrUnknownUser
		}

		return model.User{}, err
	}

	return res, err
}
