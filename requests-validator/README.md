# Requests Validator

Automatically validates your requests. Helps to incapsulate requests validation logic.

## Install

```bash
go get github.com/mlanin/iris-middlewares/requests-validator
```

## About

By Default it highly utilizes [go-apierr](https://github.com/mlanin/go-apierr) package, so for the comfort usage better be used in link with the [API Errors Handler](../apierr-handler/README.md) middleware from this repo.

Validation logic and errors handling can be overridden.

* Validates your requests by [ozzo-validation](https://github.com/go-ozzo/ozzo-validation).
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
import (
  "github.com/kataras/iris"
  "github.com/mlanin/go-apierr"
  handler "github.com/mlanin/iris-middlewares/apierr-handler"
  validator "github.com/mlanin/iris-middlewares/requests-validator"
)

// News model.
type News struct {
	Text string `json:"text"`
}

// PostNews request.
type PostNews struct {
	validator.HTTPRequest
	News
}

// Type of the request. "json" or "form".
// Used to resolve field name from the request struct's tags.
func (r *PostAPINews) Type() string {
	return "json"
}

// Validate request.
// As you see, validation logic can be anything you want.
func (r *PostNews) Validate(ctx *iris.Context) error {
  // Parse JSON in request.
	if err := ctx.ReadJSON(&r.News); err != nil {
		panic(apierr.BadRequest)
	}

  // Validate fields.
	return validation.StructRules{}.
		Add("Text", validation.Required).
		Validate(r)
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

	// Import middlewares.
  iris.Use(errorsHandler)
  iris.Use(rv)

  // Place rv.ValidateRequest with you request struct right before main handler.
	iris.Post("/news", rv.ValidateRequest(&PostNews{}), func(ctx *iris.Context) {
    // If request is valid, it will be stored by "request" key in the context.
    request := ctx.Get("request").(*PostNews)

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
		APIHandler: func(err error, ctx *iris.Context) {
			panic(apierr.BadRequest)
		},
		WebHandler: func(err error, ctx *iris.Context) {
			ctx.Text(200, "Error")
		},
	})

}
```

# Working example

Full working example can be found in [API Boilerplate](https://github.com/mlanin/go-api-biolerplate)
