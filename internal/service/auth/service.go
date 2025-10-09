package auth

import (
	"context"
	"errors"

	"github.com/ex-rate/logger"
)

type Service struct {
	secretKey []byte
	logger    *logger.Logger
}

type AuthOption func(*Service)

func WithSecretKey(secretKey []byte) AuthOption {
	return func(a *Service) {
		a.secretKey = secretKey
	}
}

func WithLogger(logger *logger.Logger) AuthOption {
	return func(a *Service) {
		a.logger = logger
	}
}

func New(opts ...AuthOption) (*Service, error) {
	auth := &Service{}

	for _, opt := range opts {
		opt(auth)
	}

	if len(auth.secretKey) == 0 {
		return nil, errors.New("secret key is required")
	}

	if auth.logger == nil {
		return nil, errors.New("logger is nil")
	}

	auth.logger.Info("auth service initialized")

	return auth, nil
}

func (s *Service) Stop(_ context.Context) error {
	return nil
}
