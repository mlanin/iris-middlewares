# Requests Validator

Automatically validates your requests. Helps to incapsulate requests validation logic.

## Install

```bash
go get github.com/mlanin/iris-middlewares/requests-validator
```

## About

By Default it highly utilizes [go-apierr](https://github.com/mlanin/go-apierr)] package, so for the comfort usage better be used in link with the apierr-handler middleware from this repo.

* Validate your requests by [ozzo-validation](https://github.com/go-ozzo/ozzo-validation).
* Send apierr.ValiationFailed error with set of `field-name` - `error-message`.
* Save valid request to the context to use its data further.

Validation logic and errors handling can be overridden.

## Usage

```go
import (
  "github.com/kataras/iris"
  "github.com/mlanin/go-apierr"
  handler "github.com/mlanin/iris-middlewares/apierr-handler"
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

  // Enable APIErrors handler middleware.
  errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})
  iris.Use(errorsHandler)

  // Place validator.ValidateRequest with you request struct right before main handler.
	iris.Post("/news", validator.ValidateRequest(&PostNews{}), func(ctx *iris.Context) {
    // If request is valid, it will be stored by "request" key in the context.
    request := ctx.Get("request").(*PostNews)

		ctx.Text(200, request.Text)
	})

}
```

## Override handler

You can override error handing logic by passing your own handler to the validation constructor.

```go
func main() {

  // Enable APIErrors handler middleware.
  errorsHandler := handler.New(handler.Config{
		EnvGetter: func() string {
			return "production"
		},
		DebugGetter: func() bool {
			return false
		},
	})
  iris.Use(errorsHandler)

  // Create new validator instance and pass your own handler.
	requestsValidator := validator.New(func(err error) {
		panic(apierr.BadRequest)
	})

  // The same logic, but use ValidateRequest method of the instance.
	iris.Post("/news", requestsValidator.ValidateRequest(&PostNews{}), func(ctx *iris.Context) {
		ctx.Text(200, "Done")
	})

}
```
