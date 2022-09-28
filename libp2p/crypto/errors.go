package crypto

import "errors"

// ErrNilPrivateKey signals that a nil private key was provided
var ErrNilPrivateKey = errors.New("nil private key")
