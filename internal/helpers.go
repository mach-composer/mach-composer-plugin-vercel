package internal

import (
	"reflect"
	"text/template"
)

var templateFunctions = template.FuncMap{
	"last": func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	},
}
