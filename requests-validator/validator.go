package validator

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kataras/iris"
	"github.com/mlanin/go-apierr"
	"github.com/serenize/snaker"
)

// HTTPRequest interface.
type HTTPRequest interface {
	Validate(ctx *iris.Context) error
}

// RequestValidator for http requests.
type RequestValidator struct {
	Request HTTPRequest
	Handler func(err error)
}

// New middleware.
func New(handler func(err error)) *RequestValidator {
	validator := &RequestValidator{
		Handler: handler,
	}

	return validator
}

// ValidateRequest helper function to make validator.
func ValidateRequest(request HTTPRequest) iris.HandlerFunc {
	validator := &RequestValidator{
		Request: request,
	}

	validator.Handler = func(err error) {
		panic(validator.makeAPIError(err))
	}

	return validator.Serve
}

// ValidateRequest helper function to make validator.
func (rv *RequestValidator) ValidateRequest(request HTTPRequest) iris.HandlerFunc {
	rv.Request = request

	return rv.Serve
}

// Serve the middleware.
func (rv *RequestValidator) Serve(ctx *iris.Context) {
	err := rv.Request.Validate(ctx)

	if err != nil {
		rv.Handler(err)
	}

	ctx.Set("request", rv.Request)
	ctx.Next()
}

// Prepare api validation error.
func (rv *RequestValidator) makeAPIError(err error) *apierr.APIError {
	errors := []apierr.ValidationError{}

	// Convert error to validation errors and append it to the meta tag.
	fails := err.(validation.Errors)
	for field, message := range fails {
		errors = append(errors, apierr.ValidationError{
			Field:   snaker.CamelToSnake(field), // Convert field name to snake_case
			Message: UcFirst(message.Error()),   // Upper case first letter of the message
		})
	}

	// Create new api error and attach meta with errors.
	fail := *apierr.ValiationFailed
	fail.AddMeta(&apierr.ValidationErrors{
		Errors: errors,
	})

	return &fail
}
