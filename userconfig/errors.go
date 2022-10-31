package userconfig

import "errors"

var (
	// ErrMissingConfig represents the error
	// returned when a resolver cannot resolve config.
	ErrMissingConfig = errors.New("ErrMissingConfig")
)
