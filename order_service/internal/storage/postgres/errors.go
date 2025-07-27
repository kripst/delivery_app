package postgres

import "errors"

var (
    ErrUnknown    = errors.New("unknown error")
    ErrNotFound  = errors.New("not found")
    ErrInvalid   = errors.New("invalid input")
)