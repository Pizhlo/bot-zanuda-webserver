package auth

import "errors"

type AuthService struct {
	secretKey []byte
}

func New(cfg *config) (*AuthService, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	return &AuthService{secretKey: cfg.secretKey}, nil
}
