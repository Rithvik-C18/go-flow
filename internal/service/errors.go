package service

import "errors"

var (
	ErrValidation        = errors.New("validation failed")
	ErrInvalidCredential = errors.New("invalid credentials")
)
