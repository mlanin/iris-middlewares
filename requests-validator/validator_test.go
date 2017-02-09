package validator_test

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/kataras/iris"
	"github.com/kataras/iris/httptest"
	apierr "github.com/mlanin/go-apierr"
	handler "github.com/mlanin/iris-middlewares/apierr-handler"
	validator "github.com/mlanin/iris-middlewares/requests-validator"
)

type News struct {
	Text string `json:"text" form:"text"`
}

// PostAPINews request.
type PostAPINews struct {
	validator.HTTPRequest
	News
}

// Type of the request.
func (r *PostAPINews) Type() string {
	return "json"
}

// Validate request.
func (r *PostAPINews) Validate(ctx *iris.Context) error {
	if err := ctx.ReadJSON(&r.News); err != nil {
		panic(err.Error())
	}

	return validation.StructRules{}.
		Add("Text", validation.Required).
		Validate(r)
}

// PostWebNews request.
type PostWebNews struct {
	validator.HTTPRequest
	News
}

// Type of the request.
func (r *PostWebNews) Type() string {
	return "form"
}

// Validate request.
func (r *PostWebNews) Validate(ctx *iris.Context) error {
	r.Text = ctx.PostValue("text")

	return validation.StructRules{}.
		Add("Text", validation.Required).
		Validate(r)
}

func TestItThrowsAPIError(t *testing.T) {
	api := iris.New()
	defer api.Close()

	rv := validator.New(validator.Config{})
	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)
	api.Use(rv)

	api.Post("/news", rv.ValidateRequest(&PostAPINews{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

	schema := `{
		"type": "object",
		"properties": {
			"error": {
				"type": "object",
				"properties": {
					"id":        {"type": "string"},
					"message":   {"type": "string"}
				},
				"required": ["id", "message"]
			},
			"meta": {
				"type": "object",
				"properties": {
					"errors":  {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "field":     {"type": "string"},
                "message":   {"type": "string"}
              },
              "required": ["field", "message"]
            }
          }
				},
				"required": ["errors"]
			}
		},
		"required": ["error", "meta"]
	}`

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithHeader("Accept", "application/json").WithJSON(map[string]interface{}{"foo": 123}).
		Expect().
		Status(iris.StatusUnprocessableEntity).
		JSON().Schema(schema)
}

func TestItHandlesWebReqest(t *testing.T) {
	api := iris.New()
	defer api.Close()

	rv := validator.New(validator.Config{})
	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)
	api.Use(rv)

	api.Get("/foo", func(ctx *iris.Context) {
		ctx.Text(200, "Foo")
	})
	api.Post("/news", rv.ValidateRequest(&PostWebNews{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true), httptest.Debug(true))
	e.POST("/news").WithHeader("Referer", "/foo").
		Expect().
		Status(iris.StatusOK).
		Body().Equal("Foo")
}

func TestItPassesIfEverythingIsOk(t *testing.T) {
	api := iris.New()
	defer api.Close()

	rv := validator.New(validator.Config{})
	errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	api.Use(errorsHandler)
	api.Use(rv)

	api.Post("/news", rv.ValidateRequest(&PostAPINews{}), func(ctx *iris.Context) {
		request := ctx.Get("request").(*PostAPINews)

		ctx.Text(200, request.Text)
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithJSON(&News{Text: "Foo bar"}).
		Expect().
		Status(iris.StatusOK).
		Body().Equal("Foo bar")
}

func TestHandlerOverride(t *testing.T) {
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

	rv := validator.New(validator.Config{
		APIHandler: func(err error, ctx *iris.Context) {
			panic(apierr.BadRequest)
		},
	})

	api.Use(errorsHandler)
	api.Use(rv)

	api.Post("/news", rv.ValidateRequest(&PostAPINews{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithHeader("Accept", "application/json").WithJSON(map[string]interface{}{"foo": 123}).
		Expect().
		Status(iris.StatusBadRequest)
}
