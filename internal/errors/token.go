package errors

import "errors"

var (
	ErrTokenExpired             = errors.New("token expired")
	ErrNoPayloadInToken         = errors.New("payload in token not found")
	ErrUserNotFoundInPayload    = errors.New("user not found in payload")
	ErrExpiredNotFoundInPayload = errors.New("expired not found in payload")
)
