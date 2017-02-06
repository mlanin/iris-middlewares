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

// PostNews request.
type PostNews struct {
	validator.HTTPRequest
	News
}

// Validate request.
func (r *PostNews) Validate(ctx *iris.Context) error {
	if err := ctx.ReadJSON(&r.News); err != nil {
		panic(err.Error())
	}

	return validation.StructRules{}.
		Add("Text", validation.Required).
		Validate(r)
}

func TestItThrowsError(t *testing.T) {
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

	api.Post("/news", validator.ValidateRequest(&PostNews{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

	schema := `{
		"type": "object",
		"properties": {
			"id":        {"type": "string"},
			"message":   {"type": "string"},
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
		"required": ["id", "message", "meta"]
	}`

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithText("{}").
		Expect().
		Status(iris.StatusUnprocessableEntity).
		JSON().Schema(schema)
}

func TestItPassesIfEverythingOk(t *testing.T) {
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

	api.Post("/news", validator.ValidateRequest(&PostNews{}), func(ctx *iris.Context) {
		request := ctx.Get("request").(*PostNews)

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

	requestsValidator := validator.New(func(err error) {
		panic(apierr.BadRequest)
	})

	api.Use(errorsHandler)

	api.Post("/news", requestsValidator.ValidateRequest(&PostNews{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

	e := httptest.New(api, t, httptest.ExplicitURL(true))
	e.POST("/news").WithText("{}").
		Expect().
		Status(iris.StatusBadRequest)
}
