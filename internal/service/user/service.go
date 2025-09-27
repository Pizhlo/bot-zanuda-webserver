package user

import (
	"context"
	"errors"
	api_errors "webserver/internal/errors"
	"webserver/internal/model"

	"github.com/ex-rate/logger"
)

type Service struct {
	repo   userRepo
	cache  userCache
	logger *logger.Logger
}

//go:generate mockgen -source ./service.go -destination=./mocks/user_srv.go -package=mocks
type userRepo interface {
	GetUser(ctx context.Context, tgID int64) (model.User, error)
	CheckUser(ctx context.Context, tgID int64) (bool, error)
}
type userCache interface {
	GetUser(ctx context.Context, tgID int64) (model.User, error)
	CheckUser(ctx context.Context, tgID int64) (bool, error)
}

type UserOption func(*Service)

func New(opts ...UserOption) (*Service, error) {
	user := &Service{}

	for _, opt := range opts {
		opt(user)
	}

	if user.repo == nil {
		return nil, errors.New("repo is nil")
	}

	if user.cache == nil {
		return nil, errors.New("cache is nil")
	}
	if user.logger == nil {
		return nil, errors.New("logger is nil")
	}

	user.logger.Info("user service initialized")

	return user, nil
}

func WithRepo(repo userRepo) UserOption {
	return func(u *Service) {
		u.repo = repo
	}
}

func WithCache(cache userCache) UserOption {
	return func(u *Service) {
		u.cache = cache
	}
}

func WithLogger(logger *logger.Logger) UserOption {
	return func(u *Service) {
		u.logger = logger
	}
}

// CheckUser проверяет существование пользователя по tgID
func (s *Service) CheckUser(ctx context.Context, tgID int64) (bool, error) {
	// проверяем в кэшэ
	exists, err := s.cache.CheckUser(ctx, tgID)
	// если ошибка не связана с отсутствием пользователя, возвращаем её
	if err != nil && !errors.Is(err, api_errors.ErrUnknownUser) {
		return false, err
	}
	// если пользователь найден в кэше, возвращаем true
	if err == nil && exists {
		return true, nil
	}

	// проверяем в БД
	exists, err = s.repo.CheckUser(ctx, tgID)
	if err != nil {
		return false, err
	}
	return exists, nil
}
