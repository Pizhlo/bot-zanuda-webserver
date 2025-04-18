package user

import (
	"context"
	"errors"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"
)

type User struct {
	repo  userRepo
	cache userCache
}

//go:generate mockgen -source ./user.go -destination=../../../mocks/user_srv.go -package=mocks
type userRepo interface {
	GetUser(tgID int64) (model.User, error)
}
type userCache interface {
	GetUser(ctx context.Context, tgID int64) (model.User, error)
}

func New(repo userRepo, cache userCache) *User {
	return &User{repo: repo, cache: cache}
}

// CheckUser проверяет существование пользователя по tgID
func (s *User) CheckUser(ctx context.Context, tgID int64) error {
	// проверяем в кэшэ
	_, err := s.cache.GetUser(ctx, tgID)
	if err != nil {
		if !errors.Is(err, api_errors.ErrUnknownUser) {
			return err
		}
	}

	if err == nil {
		return nil
	}

	// в кэшэ не найдено - проверяем в БД
	_, err = s.repo.GetUser(tgID)
	return err
}
