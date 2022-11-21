package transaction

import (
	"errors"
)

var (
	// ErrInvalidTransactionBuilderRequest is the error returned when a transaction builder request is invalid
	ErrInvalidTransactionBuilderRequest = errors.New("invalid transaction builder request")
)
