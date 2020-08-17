package browserk

import "errors"

var (
	// ErrInjectionTimeout happened
	ErrInjectionTimeout       = errors.New("injection timed out")
	ErrEmptyInjectionResponse = errors.New("injection body was empty")
)
