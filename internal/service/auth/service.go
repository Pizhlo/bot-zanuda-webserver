package auth

import "errors"

type AuthService struct {
	secretKey []byte
}

type AuthOption func(*AuthService)

func WithSecretKey(secretKey []byte) AuthOption {
	return func(a *AuthService) {
		a.secretKey = secretKey
	}
}

func New(opts ...AuthOption) (*AuthService, error) {
	auth := &AuthService{}

	for _, opt := range opts {
		opt(auth)
	}

	if len(auth.secretKey) == 0 {
		return nil, errors.New("secret key is required")
	}

	return auth, nil
}
