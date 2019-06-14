package funclown

import "errors"

var (
	// ErrResourceNotFound is kind of errors for missing database resource.
	ErrResourceNotFound = errors.New("resource not found")
)
