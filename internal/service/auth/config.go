package auth

import "errors"

type config struct {
	secretKey []byte
}

func NewConfig(secetKey []byte) (*config, error) {
	if len(secetKey) == 0 {
		return nil, errors.New("secret key not provided")
	}

	return &config{secretKey: secetKey}, nil
}
