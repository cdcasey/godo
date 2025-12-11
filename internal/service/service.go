package service

import (
	"errors"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrLastAdmin = errors.New("last admin")
)
