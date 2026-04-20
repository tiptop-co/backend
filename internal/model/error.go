package model

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
)
