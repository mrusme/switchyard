package errs

import (
	"errors"
)

var (
	ErrConfigTypeUnsupported error = errors.New(
		"The configuration type is unsupported, use a file:// URL or a path",
	)
)
