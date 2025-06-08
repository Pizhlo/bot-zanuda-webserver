package auth

import "errors"

type Service struct {
	secretKey []byte
}

type AuthOption func(*Service)

func WithSecretKey(secretKey []byte) AuthOption {
	return func(a *Service) {
		a.secretKey = secretKey
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

	return auth, nil
}
