package common

import "errors"

var (
	ErrTxnLockFail = errors.New("attempt to lock, but failed")
	ErrNotFoundIP  = errors.New("unable to get local ipv4")
)
