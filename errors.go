package rcs

import "errors"

var (
	ErrEmptyId       = errors.New("empty id")
	ErrRevisionEmpty = errors.New("revision empty")
	ErrDateParse     = errors.New("unable to parse date")
	ErrUnknownToken  = errors.New("unknown token")
)
