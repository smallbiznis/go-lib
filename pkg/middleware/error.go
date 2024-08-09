package middleware

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/smallbiznis/go-lib/pkg/errors"
)

func HandleError(translate ut.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if err := c.Errors.Last(); err != nil {
			c.JSON(
				validationError(err.Err, translate),
			)
		}
	}
}

func validationError(err error, translate ut.Translator) (code int, obj any) {
	code = 500
	obj = gin.H{
		"error": errors.InternalServerError("InternalServerError", err.Error()),
	}

	// Handle error io.EOF request body empty
	if err == io.EOF {
		code = 400
		obj = gin.H{
			"error": errors.BadRequest("InvalidRequest", "request can't be empty"),
		}
		return
	}

	// Handle error *json.SyntaxError
	if _, ok := err.(*json.SyntaxError); ok {
		code = 400
		obj = gin.H{
			"error": errors.BadRequest("InvalidRequest", err.Error()),
		}
		return
	}

	// Handle error *validator.fieldError
	if e, ok := err.(validator.FieldError); ok {
		code = 400
		msg := fmt.Errorf(e.Translate(translate)).Error()
		newErr := errors.BadRequest("InvalidRequest", msg)
		obj = gin.H{
			"error": newErr,
		}
		return
	}

	// Handle error uuid.isInvalidLength
	if uuid.IsInvalidLengthError(err) {
		code = 400
		obj = gin.H{
			"error": err,
		}
		return
	}

	// Handle error business logic
	if _, ok := err.(errors.Error); ok {
		code = 400
		obj = gin.H{
			"error": err,
		}
		return
	}

	return
}
