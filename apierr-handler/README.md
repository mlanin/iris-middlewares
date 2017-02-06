# API Errors Handler

Handles errors sent by handy [go-apierr](https://github.com/mlanin/go-apierr) package.

## Install

```bash
go get github.com/mlanin/iris-middlewares/apierr-handler
```

## Can

* Recover APIErrors and send them right to the user.
* Handles APIError's context and trace options.

Also it can transform all uncached panic errors into InternalServerError and saves them to logs.

If your environment is set to `production`, user will never see your panics.
If you want to see them, set your env to anything else like `develop` by `EnvGetter` option.

By default only InternalServerErrors are saved to log. If you want to save every error,
set debug mode to `true` by `DebugGetter`.

## Usage

```go
import (
  "github.com/kataras/iris"
  handler "github.com/mlanin/iris-middlewares/apierr-handler"
)

func main() {
  errorsHandler := handler.New(handler.Config{
    // Set your environment.
    EnvGetter: func() string {
      return "production"
    },
    // Set your debug mode.
    DebugGetter: func() bool {
      return false
    },
  })

  iris.Use(errorsHandler)
}
```

# Working example

Full working example can be found in [API Boilerplate](https://github.com/mlanin/go-api-biolerplate)
