package types

import "github.com/cockroachdb/errors"

var (
	ErrInvalidCapacity  = errors.New("invalid resource capacity")
	ErrInvalidGPUMap    = errors.New("invalid gpu map")
	ErrInvalidGPU       = errors.New("invalid gpu")
	ErrInvalidMergeType = errors.New("invalid merge type")
)
