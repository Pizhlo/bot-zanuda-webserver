package user

import (
	"webserver/internal/model"
	"webserver/internal/service/storage/postgres/space"
)

type User struct {
	repo  userRepo
	cache cache
}

type userRepo interface {
	GetUser(tgID int64) (model.User, error)
}
type cache interface {
	GetUser(tgID int64) (model.User, error)
}

func New(repo userRepo, cache cache) *User {
	return &User{repo: repo, cache: cache}
}

// CheckUser проверяет существование пользователя по tgID
func (s *User) CheckUser(tgID int64) error {
	// проверяем в кэшэ
	if _, err := s.cache.GetUser(tgID); err == nil {
		return nil
	}

	// в кэшэ не найдено - проверяем в БД
	_, err := s.repo.GetUser(tgID)
	if err == nil {
		return nil
	}

	return space.ErrUnknownUser
}
