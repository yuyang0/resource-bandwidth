package types

import "github.com/cockroachdb/errors"

var (
	ErrInvalidCapacity  = errors.New("invalid resource capacity")
	ErrInvalidBandwidth = errors.New("invalid bandwidth")
)
