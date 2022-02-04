package util

import (
	"errors"
)

var (
	// ErrInvalidNonce is the error returned when a nonce is invalid
	ErrInvalidNonce = errors.New("invalid nonce")
)
