package models

import "errors"

var (
	ErrRecordNotFound            = errors.New("record not found")
	ErrUniqueConstraintViolation = errors.New("unique key constraint violated")
)
