package user

import "webserver/internal/model"

func (db *userRepo) GetUser(tgID int64) (model.User, error) {
	return model.User{}, nil
}
