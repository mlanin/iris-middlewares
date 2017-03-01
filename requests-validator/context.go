package validator

import (
	"reflect"
	"strings"
)

const (
	jsonRequest   = "json"
	xmlRequest    = "xml"
	formRequest   = "form"
	queryRequest  = "query"
	paramsRequest = "params"
)

// Context to store request data.
type Context struct {
	Request HTTPRequest
	Name    string
	Type    string
	Errors  error
}

// NewContext constructor.
func NewContext(request HTTPRequest) *Context {
	context := &Context{
		Request: request,
	}

	context.Name = reflect.TypeOf(request).Elem().String()

	switch true {
	case strings.HasSuffix(context.Name, "JSON"):
		context.Type = jsonRequest
	case strings.HasSuffix(context.Name, "XML"):
		context.Type = xmlRequest
	case strings.HasSuffix(context.Name, "Form"):
		context.Type = formRequest
	case strings.HasSuffix(context.Name, "Query"):
		context.Type = queryRequest
	case strings.HasSuffix(context.Name, "Params"):
		context.Type = paramsRequest
	}

	return context
}
