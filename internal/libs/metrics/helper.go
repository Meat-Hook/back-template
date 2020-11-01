package metrics

import (
	"reflect"
	"runtime"
	"strings"
)

// Labels.
const (
	ResourceLabel = "resource"
	MethodLabel   = "method"
	CodeLabel     = "code"
)

// MethodsOf require pointer to interface (e.g.: new(app.Appl)) and
// returns all it methods.
func MethodsOf(v interface{}) []string {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		panic("require pointer to interface")
	}

	methods := make([]string, typ.NumMethod())
	for i := 0; i < typ.NumMethod(); i++ {
		methods[i] = typ.Method(i).Name
	}

	return methods
}

// CallerMethodName returns caller's method name for given stack depth.
func CallerMethodName() string {
	pc, _, _, _ := runtime.Caller(2)
	names := strings.Split(runtime.FuncForPC(pc).Name(), ".")

	return names[len(names)-1]
}

// CallerPkg returns caller's package name (from path) for given stack depth.
func CallerPkg() string {
	pc, _, _, _ := runtime.Caller(2)
	names := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	pkg := names[len(names)-1]

	return pkg[:strings.Index(pkg, ".")]
}
