package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Service) CheckToken(authHeader string) (*jwt.Token, error) {
	s.logger.Debug("checking token")

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := s.ParseToken(tokenString)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	payload, ok := s.GetPayload(token)
	if !ok {
		return nil, errors.New("no payload found")
	}

	expiredAny, ok := payload["expired"]
	if !ok {
		return nil, errors.New("not found 'expired' key")
	}

	expired, ok := expiredAny.(float64)
	if !ok {
		return nil, errors.New("failed to convert interface to int64")
	}

	if time.Now().Unix() > int64(expired) {
		return nil, errors.New("token is expired")
	}

	return token, nil
}

func (s *Service) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})
}

func (s *Service) GetPayload(token *jwt.Token) (jwt.MapClaims, bool) {
	if token == nil {
		return nil, false
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}

	return payload, true
}
