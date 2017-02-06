package validator_test

import (
	"testing"

	validator "github.com/mlanin/iris-middlewares/requests-validator"
)

type testpair struct {
	value  string
	result string
}

var ucfirstTests = []testpair{
	{"foo", "Foo"},
	{"foo bar", "Foo bar"},
	{"Foo bar", "Foo bar"},
	{"", ""},
}

func TestUcFirst(t *testing.T) {
	for _, pair := range ucfirstTests {
		v := validator.UcFirst(pair.value)
		if v != pair.result {
			t.Error(
				"For", pair.value,
				"expected", pair.result,
				"got", v,
			)
		}
	}
}
