package domain

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrForbidden      = errors.New("user does not have permission")
	ErrBadRequest     = errors.New("invalid input")
	ErrConflict       = errors.New("resource conflict (e.g., duplicate)")
	ErrUnauthorized   = errors.New("authentication required or failed")
	ErrInternalServer = errors.New("internal server error")
	ErrValidation     = errors.New("validation failed")
)
