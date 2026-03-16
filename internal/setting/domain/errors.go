package domain

import "errors"

var (
	ErrSettingNotFound = errors.New("setting not found")
	ErrInvalidKey      = errors.New("invalid setting key")
	ErrEmptyValue      = errors.New("setting value cannot be empty")
)
