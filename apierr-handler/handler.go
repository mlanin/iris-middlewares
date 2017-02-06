package handler

import (
	"errors"
	"fmt"
	"net/http"
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
		if err := recover(); err != nil {
			fail := h.convertToAPIError(err)

			if h.needToReport(fail) {
				ctx.Log("[apierr.APIError] %+v [%+v]", err, fail.Context)
				if h.needToAddTrace(fail) {
					ctx.Log("[Trace] %s", debug.Stack())
				}
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
		ID:           "internal_server_error",
		Message:      err.Error(),
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