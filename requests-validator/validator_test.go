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
	Text string `json:"text"`
}

type PostNewsJSON struct {
	Text string `json:"text"`
}

func (r *PostNewsJSON) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Text, validation.Required),
	)
}

type PostNewsForm struct {
	Text string `form:"text"`
}

func (r *PostNewsForm) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Text, validation.Required),
	)
}

// PostNewsQuery request.
type PostNewsQuery struct {
	Text    string `query:"text"`
	Integer int    `query:"int"`
}

// Validate request.
func (r *PostNewsQuery) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Text, validation.Required),
	)
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

	api.Post("/news", rv.ValidateRequest(&PostNewsJSON{}), func(ctx *iris.Context) {
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
	api.Post("/news", rv.ValidateRequest(&PostNewsForm{}), func(ctx *iris.Context) {
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

	api.Post("/news", rv.ValidateRequest(&PostNewsJSON{}), func(ctx *iris.Context) {
		request := ctx.Get("validator_test.PostNewsJSON").(*PostNewsJSON)

		ctx.Text(200, request.Text)
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithJSON(&News{Text: "Foo bar"}).
		Expect().
		Status(iris.StatusOK).
		Body().Equal("Foo bar")
}

func TestItPassesQueryIfEverythingIsOk(t *testing.T) {
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

	api.Get("/news", rv.ValidateRequest(&PostNewsQuery{}), func(ctx *iris.Context) {
		request := ctx.Get("validator_test.PostNewsQuery").(*PostNewsQuery)

		ctx.Text(200, request.Text)
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.GET("/news").WithHeader("Accept", "application/json").WithQuery("text", "Foo bar").WithQuery("int", 123).
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
		APIHandler: func(context *validator.Context, ctx *iris.Context) {
			panic(apierr.NotFound)
		},
	})

	api.Use(errorsHandler)
	api.Use(rv)

	api.Post("/news", rv.ValidateRequest(&PostNewsJSON{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithHeader("Accept", "application/json").WithJSON(map[string]interface{}{"foo": 123}).
		Expect().
		Status(iris.StatusNotFound)
}
