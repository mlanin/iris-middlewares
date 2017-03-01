package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"git.acronis.com/ci/ci-2x-ipn/app/support/utils"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kataras/iris"
	"github.com/mlanin/go-apierr"
)

const (
	jsonTag   = jsonRequest
	xmlTag    = xmlRequest
	formTag   = formRequest
	queryTag  = queryRequest
	paramsTag = paramsRequest
)

// HTTPRequest interface.
type HTTPRequest interface {
	// Validate request.
	Validate() error
}

// Config for the middleware.
type Config struct {
	// Handle error for common http request.
	WebHandler func(context *Context, ctx *iris.Context)
	// Handle error for API calls.
	APIHandler func(context *Context, ctx *iris.Context)
	// Handle unmarshal request error.
	BadRequestHandler func(context *Context, ctx *iris.Context)
}

// RequestsValidator for http requests.
type RequestsValidator struct {
	Config
}

// New middleware constructor.
func New(config Config) *RequestsValidator {
	validator := &RequestsValidator{
		config,
	}

	if validator.APIHandler == nil {
		validator.APIHandler = validator.sendAPIError
	}
	if validator.WebHandler == nil {
		validator.WebHandler = validator.sendWebError
	}
	if validator.BadRequestHandler == nil {
		validator.BadRequestHandler = validator.sendBadRequest
	}

	return validator
}

// Serve the middleware.
func (rv *RequestsValidator) Serve(ctx *iris.Context) {
	ctx.Next()

	// Save current url.
	rv.storeCurrentURL(ctx)
}

// ValidateRequest helper function to make validator.
func (rv *RequestsValidator) ValidateRequest(request HTTPRequest) iris.HandlerFunc {
	context := NewContext(request)

	return func(ctx *iris.Context) {
		context.Errors = rv.populateRequest(context, ctx)
		if context.Errors != nil {
			rv.BadRequestHandler(context, ctx)
		}

		context.Errors = context.Request.Validate()

		if context.Errors == nil {
			// Save request to use it futher in controller.
			ctx.Set(context.Name, context.Request)

			// Switch to next handler.
			ctx.Next()
			return
		}

		// Convert errors to validation fails and send them to the user.
		if rv.wantsJSON(ctx) {
			rv.APIHandler(context, ctx)
		} else {
			rv.WebHandler(context, ctx)
		}
	}
}

// Populate Request with data.
func (rv *RequestsValidator) populateRequest(context *Context, ctx *iris.Context) error {
	switch context.Type {
	case jsonRequest:
		return ctx.ReadJSON(&context.Request)
	case xmlRequest:
		return ctx.ReadXML(&context.Request)
	case formRequest:
		return ctx.ReadForm(&context.Request)
	case queryRequest:
		return rv.populateFromQuery(context.Request, ctx)
	case paramsRequest:
		return rv.populateFromParams(context.Request, ctx)
	}

	return nil
}

// Populate requst from query.
func (rv *RequestsValidator) populateFromQuery(request HTTPRequest, ctx *iris.Context) error {
	return rv.populateFromSource(request, ctx.URLParam)
}

// Populate request from URL params.
func (rv *RequestsValidator) populateFromParams(request HTTPRequest, ctx *iris.Context) error {
	return rv.populateFromSource(request, ctx.Param)
}

// Populate Request with data from custom source.
func (rv *RequestsValidator) populateFromSource(request HTTPRequest, source func(key string) string) error {
	var queryName string
	var queryValue string

	v := reflect.ValueOf(request).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldType := t.Field(i)

		if !fieldType.Anonymous && fieldValue.IsValid() && fieldValue.CanSet() {
			tag := fieldType.Tag.Get(queryTag)

			if tag != "" {
				queryName = tag
			} else {
				queryName = fieldType.Name
			}

			queryValue = source(queryName)

			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(queryValue)
			case reflect.Int:
				integer, err := strconv.ParseInt(queryValue, 10, 64)
				if err != nil {
					return fmt.Errorf("Expected integer for query field %s, but found '%s' instead.", fieldType.Name, queryValue)
				}

				fieldValue.SetInt(integer)
			case reflect.Bool:
				boolean, err := strconv.ParseBool(queryValue)
				if err != nil {
					return fmt.Errorf("Expected boolean for query field %s, but found '%s' instead.", fieldType.Name, queryValue)
				}

				fieldValue.SetBool(boolean)
			default:
				return fmt.Errorf("Request can obtain only string, integer or boolean. %s field found with %s type.", fieldType.Name, fieldValue.Kind())
			}
		}
	}

	return nil
}

// If user want's only JSON.
func (rv *RequestsValidator) wantsJSON(ctx *iris.Context) bool {
	return strings.Contains(ctx.RequestHeader("accept"), "application/json")
}

// Add validation errors to flash and send back 302 redirect.
func (rv *RequestsValidator) sendWebError(context *Context, ctx *iris.Context) {
	errors := rv.convertErrors(context)

	ctx.Session().SetFlash("_errors", errors)
	ctx.Session().SetFlash("_old_input", context.Request)

	rv.redirectBack(ctx)
}

// Send API validation error.
func (rv *RequestsValidator) sendAPIError(context *Context, ctx *iris.Context) {
	errors := rv.convertErrors(context)

	// Create new api error and attach meta with errors.
	fail := *apierr.ValiationFailed
	fail.AddMeta(&apierr.ValidationErrors{
		Errors: errors,
	})
	fail.AddContext(errors)

	panic(&fail)
}

// Default bad request callback.
func (rv *RequestsValidator) sendBadRequest(context *Context, ctx *iris.Context) {
	panic(context.Errors)
}

// Store current url to reuse it for
func (rv *RequestsValidator) storeCurrentURL(ctx *iris.Context) {
	if ctx.Method() == "GET" && !ctx.IsAjax() && !rv.wantsJSON(ctx) {
		ctx.Session().Set("_previous_url", ctx.Request.URL.String())
	}
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

// Convert errors to field - message format.
func (rv *RequestsValidator) convertErrors(context *Context) []apierr.ValidationError {
	fails := context.Errors.(validation.Errors)
	errors := []apierr.ValidationError{}

	reflection := reflect.ValueOf(context.Request).Elem().Type()

	for field, message := range fails {
		errors = append(errors, apierr.ValidationError{
			Field:   rv.normalizeFieldName(reflection, context.Type, field),
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
	return utils.UcFirst(message.Error())
}
