package rcs

import "errors"

var (
	ErrEmptyId         = errors.New("empty id")
	ErrRevisionEmpty   = errors.New("revision empty")
	ErrDateParse       = errors.New("unable to parse date")
	ErrUnknownToken    = errors.New("unknown token")
	ErrTooManyNewLines = errors.New("too many new lines")
)

type ErrParseProperty struct {
	Property string
	Err      error
}

func (e ErrParseProperty) Error() string {
	return "expected value for " + e.Property + ": " + e.Err.Error()
}

func (e ErrParseProperty) Unwrap() error {
	return e.Err
}
