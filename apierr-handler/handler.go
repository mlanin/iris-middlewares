package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"

	"github.com/kataras/iris"
	"github.com/mlanin/go-apierr"
)

// Config for the Handler.
type Config struct {
	EnvGetter   func() string
	DebugGetter func() bool
}

// Handler for APIErrors.
type Handler struct {
	Config Config
}

// New restores the server on internal server errors (panics)
// returns the middleware
//
// is here for compatiblity
func New(cfg Config) *Handler {
	return &Handler{Config: cfg}
}

// Serve the middleware.
func (h *Handler) Serve(ctx *iris.Context) {
	defer func() {
		messages := make([]interface{}, 0)

		if err := recover(); err != nil {
			fail := h.convertToAPIError(err)

			if h.needToReport(fail) {
				messages = append(messages, fmt.Sprintf("[apierr.APIError] %+v [%+v]\n", err, fail.Context))
				if h.needToAddTrace(fail) {
					messages = append(messages, fmt.Sprintf("--> %+v\n", h.thrower()))
					messages = append(messages, string(debug.Stack()))
				}
				ctx.Log(fmt.Sprintln(messages...))
			}

			ctx.JSON(fail.HTTPCode, fail)
		}
	}()

	ctx.Next()
}

// Converts catched error to internal apierr.APIError instance.
func (h *Handler) convertToAPIError(err interface{}) *apierr.APIError {
	var fail *apierr.APIError

	switch err := err.(type) {
	case *apierr.APIError:
		fail = err
	case error:
		fail = h.NewAPIError(err.(error))
	case string:
		fail = h.NewAPIError(errors.New(err))
	default:
		fail = h.NewAPIError(fmt.Errorf("%v", err))
	}

	return fail
}

// NewAPIError makes new API error.
func (h *Handler) NewAPIError(err error) *apierr.APIError {
	// Don't show unknown error text to user when in production.
	if h.isProduction() {
		return apierr.InternalServerError
	}

	return &apierr.APIError{
		Body: apierr.Body{
			ID:      "internal_server_error",
			Message: err.Error(),
		},
		HTTPCode:     http.StatusInternalServerError,
		ShouldReport: true,
	}
}

// Check if we neer to report the error.
func (h *Handler) needToReport(fail *apierr.APIError) bool {
	return fail.WantsToBeReported() || !h.isProduction()
}

// Check if we neer to report the error.
func (h *Handler) needToAddTrace(fail *apierr.APIError) bool {
	return fail.WantsToShowTrace() || !h.isDebugOn()
}

// Check if app in production mode.
func (h *Handler) isProduction() bool {
	return h.Config.EnvGetter() == "production"
}

// Check if app in debug mode is on.
func (h *Handler) isDebugOn() bool {
	return h.Config.DebugGetter()
}

// Return a possible line, where panic was thrown.
func (h *Handler) thrower() string {
	var goSrcRegexp = regexp.MustCompile(`(mlanin/apierr-handler/.*.go)|(libexec/src/runtime)`)
	var goTestRegexp = regexp.MustCompile(`mlanin/apierr-handler/.*test.go`)

	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!goSrcRegexp.MatchString(file) || goTestRegexp.MatchString(file)) {
			return fmt.Sprintf("%v:%v", file, line)
		}
	}

	return ""
}
