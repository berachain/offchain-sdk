package eth

import "errors"

var (
	ErrAlreadyDial = errors.New("client is already dialed, please Close() before dialing again")
	ErrClosed      = errors.New("client is already closed, please Dial() before closing again")
)
