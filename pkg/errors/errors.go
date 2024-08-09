package errors

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type Error interface {
	Error() string
}

type apiError struct {
	Status  int     `json:"status"`
	Name    string  `json:"name"`
	Msg     string  `json:"message"`
	Details []gin.H `json:"details"`
}

var _ Error = new(apiError)
var _ error = new(apiError)

func (e *apiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// New
func New(code int, name, message string) error {
	return &apiError{
		Status: code,
		Name:   name,
		Msg:    message,
	}
}

// Unauthorized
func Unauthorized(name, message string) error {
	return &apiError{
		Status: 401,
		Name:   name,
		Msg:    message,
	}
}

// Forbidden
func Forbidden(name, message string) error {
	return &apiError{
		Status: 403,
		Name:   name,
		Msg:    message,
	}
}

// BadRequest
func BadRequest(name, message string) error {
	return &apiError{
		Status: 400,
		Name:   name,
		Msg:    message,
	}
}

// InternalServerError
func InternalServerError(name, message string) error {
	return &apiError{
		Status: 500,
		Name:   name,
		Msg:    message,
	}
}

// MultiError
type MultiError struct {
	Errors []Error `json:"errors"`
}

// NewMultiError
func NewMultiError() *MultiError {
	return &MultiError{
		Errors: make([]Error, 0),
	}
}

// HasError
func (e *MultiError) HasError() bool {
	return len(e.Errors) > 0
}

// Append
func (e *MultiError) Append(err ...Error) {
	e.Errors = append(e.Errors, err...)
}

// Error
func (e *MultiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}
