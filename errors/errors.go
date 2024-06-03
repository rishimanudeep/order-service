package errors

import "fmt"

type EntityNotFound struct {
	Message string
}

func (e *EntityNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

type InternalServerError struct {
	Message string
}

func (e *InternalServerError) Error() string {
	return fmt.Sprintf("internal server error: %s", e.Message)
}

type NoResponse struct {
	Message string
}

func (e *NoResponse) Error() string {
	return fmt.Sprintf("content not found in db: %s", e.Message)
}

type MissingParam struct {
	Message string
}

func (e *MissingParam) Error() string {
	return fmt.Sprintf("missing parameter: %s", e.Message)
}
