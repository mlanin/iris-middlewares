package handler_test

import (
	"errors"
	"testing"

	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	"github.com/mlanin/go-apierr"
	handler "github.com/mlanin/iris-middlewares/apierr-handler"
)

func TestItCatchesPanicWithAPIError(t *testing.T) {
	api := iris.New()
	defer api.Close()

	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)

	api.Get("/", func(ctx *iris.Context) {
		panic(apierr.NotFound)
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/").
		Expect().
		Status(iris.StatusNotFound).
		JSON().Object().ContainsKey("id").ValueEqual("id", "not_found")
}

func TestItCatchesPanicWithAPIErrorWithMeta(t *testing.T) {
	api := iris.New()
	defer api.Close()

	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)

	api.Get("/", func(ctx *iris.Context) {
		fail := *apierr.BadRequest
		fail.AddMeta(struct {
			Fail string `json:"fail"`
		}{
			Fail: "ID must be a valid integer.",
		})
		panic(&fail)
	})

	schema := `{
		"type": "object",
		"properties": {
			"id":        {"type": "string"},
			"message":   {"type": "string"},
			"meta": {
				"type": "object",
				"properties": {
					"fail":  {"type": "string"}
				},
				"required": ["fail"]
			}
		},
		"required": ["id", "message", "meta"]
	}`

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/").
		Expect().
		Status(iris.StatusBadRequest).
		JSON().Schema(schema)
}

func TestItCatchesPanicWithNativeError(t *testing.T) {
	api := iris.New()
	defer api.Close()

	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)

	api.Get("/", func(ctx *iris.Context) {
		panic(errors.New("Error"))
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/").
		Expect().
		Status(iris.StatusInternalServerError).
		JSON().Object().ContainsKey("id").ValueEqual("id", "internal_server_error")
}

func TestItCatchesPanicWithString(t *testing.T) {
	api := iris.New()
	defer api.Close()

	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)

	api.Get("/", func(ctx *iris.Context) {
		panic("Error")
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/").
		Expect().
		Status(iris.StatusInternalServerError).
		JSON().Object().ContainsKey("id").ValueEqual("id", "internal_server_error")
}

func TestItCatchesPanicWithStringForLocal(t *testing.T) {
	api := iris.New()
	defer api.Close()

	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "local"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)

	api.Get("/", func(ctx *iris.Context) {
		panic("foo")
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/").
		Expect().
		Status(iris.StatusInternalServerError).
		JSON().Object().ContainsKey("message").ValueEqual("message", "foo")
}

func TestItCatchesPanicWithAnyValue(t *testing.T) {
	api := iris.New()
	defer api.Close()

	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)

	api.Get("/", func(ctx *iris.Context) {
		panic(false)
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/").
		Expect().
		Status(iris.StatusInternalServerError).
		JSON().Object().ContainsKey("id").ValueEqual("id", "internal_server_error")
}
