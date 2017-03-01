# Requests Validator

Automatically validates your requests. Helps to incapsulate requests validation logic.

## Install

```bash
go get github.com/mlanin/iris-middlewares/requests-validator
```

## About

By Default it highly utilizes [go-apierr](https://github.com/mlanin/go-apierr) package, so for the comfort usage better be used in link with the [API Errors Handler](../apierr-handler/README.md) middleware from this repo.

Validation logic and errors handling can be overridden.

* Populates your requests from the preferred source.
* Validates requests data by [ozzo-validation(v3)](https://github.com/go-ozzo/ozzo-validation).
* Saves valid request to the context to use its data further.

### JSON API validation

On invalid request will panic with `apierr.ValidationFailed` that can be converted in 422 response with JSON like:

```json
{
	"error": {
		"id": "validation_failed",
		"message": "Validation failed.",
	},
	"meta": {
		"errors": [
      {
				"field": "body",
				"message": "Cannot be blank",
			}
    ]
	}
}
```

### Common web forms validation

On invalid requests old input and validation errors will be saved to session flash and redirected back
to the referer url or previous visited page or `/`.

## Usage

```go
package main

import (
  "github.com/kataras/iris"
  "github.com/mlanin/go-apierr"
  handler "github.com/mlanin/iris-middlewares/apierr-handler"
  validator "github.com/mlanin/iris-middlewares/requests-validator"
)

// PostNewsJSON request:
// - if name ends with JSON, request will be populated from JSON;
// - if name ends with XML, request will be populated with XML;
// - if name ends with Form, request will be populated from form data;
// - if name ends with Query, request will be populated from URL query
type PostNewsJSON struct {
	// By default fields in request will be searched by their name, but you can override it by tag:
	// - json
	// - xml
	// - form
	// - query
	Text string `json:"text"`
}

// Validate request.
func (r *PostNews) Validate() error {
	// Register validation rules.
	return validation.ValidateStruct(r,
		validation.Field(&r.Text, validation.Required),
	)
}

func main() {

	// Make RequestsValidator middleware.
	rv := validator.New(validator.Config{})

  // Make APIErrors handler middleware.
  errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})

	// Import middleware.
  iris.Use(errorsHandler)
  iris.Use(rv)

  // Place rv.ValidateRequest with you request struct right before main handler.
	iris.Post("/news", rv.ValidateRequest(&PostNewsJSON{}), func(ctx *iris.Context) {
    // If request is valid, it will be stored by request full name key in the IRIS context.
    request := ctx.Get("main.PostNewsJSON").(*PostNewsJSON)

		ctx.Text(200, request.Text)
	})

}
```

## Override handler

You can override error handing logic by passing your own handlers to the validator's constructor.

```go
func main() {

	// Make RequestsValidator middleware and override default JSON API requests errors.
	rv := validator.New(validator.Config{
		APIHandler: func(context *validator.Context, ctx *iris.Context) {
			panic(apierr.BadRequest)
		},
		WebHandler: func(context *validator.Context, ctx *iris.Context) {
			ctx.Text(200, "Error")
		},
		BadRequestHandler: func(context *validator.Context, ctx *iris.Context) {
			panic(apierr.BadRequest)
		},
	})

}
```

# Working example

Full working example can be found in [API Boilerplate](https://github.com/mlanin/go-api-biolerplate)
