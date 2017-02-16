package validator

import (
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kataras/iris"
	"github.com/mlanin/go-apierr"
)

// HTTPRequest interface.
type HTTPRequest interface {
	// Type of the request.
	Type() string
	// Validate request.
	Validate(ctx *iris.Context) error
}

// Config for the middleware.
type Config struct {
	WebHandler func(err error, ctx *iris.Context)
	APIHandler func(err error, ctx *iris.Context)
}

// RequestsValidator for http requests.
type RequestsValidator struct {
	Request HTTPRequest
	Config
}

// New middleware.
func New(config Config) *RequestsValidator {
	validator := &RequestsValidator{
		Config: config,
	}

	if validator.APIHandler == nil {
		validator.APIHandler = validator.sendAPIError
	}

	if validator.WebHandler == nil {
		validator.WebHandler = validator.sendWebError
	}

	return validator
}

// ValidateRequest helper function to make validator.
func (rv *RequestsValidator) ValidateRequest(request HTTPRequest) iris.HandlerFunc {
	rv.Request = request

	return rv.Validate
}

// Serve the middleware.
func (rv *RequestsValidator) Serve(ctx *iris.Context) {
	ctx.Next()

	// Save current url.
	rv.storeCurrentURL(ctx)
}

// Validate request.
func (rv *RequestsValidator) Validate(ctx *iris.Context) {
	err := rv.Request.Validate(ctx)

	if err == nil {
		// Save request to use it futher in controller.
		ctx.Set("request", rv.Request)

		// Switch to next handler.
		ctx.Next()
		return
	}

	// Convert errors to validation fails and send them to the user.
	if rv.wantsJSON(ctx) {
		rv.APIHandler(err, ctx)
	} else {
		rv.WebHandler(err, ctx)
	}
}

// If user want's only JSON.
func (rv *RequestsValidator) wantsJSON(ctx *iris.Context) bool {
	return ctx.RequestHeader("accept") == "application/json"
}

// Store current url to reuse it for
func (rv *RequestsValidator) storeCurrentURL(ctx *iris.Context) {
	if ctx.Method() == "GET" && !ctx.IsAjax() && !rv.wantsJSON(ctx) {
		ctx.Session().Set("_previous_url", ctx.Request.URL.String())
	}
}

// Add validation errors to flash and send back 302 redirect.
func (rv *RequestsValidator) sendWebError(err error, ctx *iris.Context) {
	errors := rv.convertErrors(err)

	ctx.Session().SetFlash("_errors", errors)
	ctx.Session().SetFlash("_old_input", rv.Request)

	rv.redirectBack(ctx)
}

// Redirect user back to show page with errors.
func (rv *RequestsValidator) redirectBack(ctx *iris.Context) {
	referer := ctx.RequestHeader("referer")
	if referer != "" {
		ctx.Redirect(referer, 302)
		return
	}

	previousURL := ctx.Session().GetFlash("_previous_url")
	if previousURL != nil {
		ctx.Redirect(previousURL.(string), 302)
		return
	}

	ctx.Redirect("/", 302)
}

// Send API validation error.
func (rv *RequestsValidator) sendAPIError(err error, ctx *iris.Context) {
	errors := rv.convertErrors(err)

	// Create new api error and attach meta with errors.
	fail := *apierr.ValiationFailed
	fail.AddMeta(&apierr.ValidationErrors{
		Errors: errors,
	})

	panic(&fail)
}

// Convert errors to field - message format.
func (rv *RequestsValidator) convertErrors(err error) []apierr.ValidationError {
	fails := err.(validation.Errors)
	errors := []apierr.ValidationError{}

	reflection := reflect.ValueOf(rv.Request).Elem().Type()
	requestType := rv.Request.Type()

	for field, message := range fails {
		errors = append(errors, apierr.ValidationError{
			Field:   rv.normalizeFieldName(reflection, requestType, field),
			Message: rv.normalizeMessage(message),
		})
	}

	return errors
}

// Convert request attribute name to normal.
func (rv *RequestsValidator) normalizeFieldName(reflection reflect.Type, requestType string, name string) string {
	field, ok := reflection.FieldByName(name)
	if ok {
		return field.Tag.Get(requestType)
	}

	return name
}

// Upper case first letter of the message.
func (rv *RequestsValidator) normalizeMessage(message error) string {
	return UcFirst(message.Error())
}
