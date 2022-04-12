package common

import "errors"

var (
	ErrTxnLockFail = errors.New("attempt to lock, but failed")
)
