package model

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidParams = errors.New("invalid parameters")
)
