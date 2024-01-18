package test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Name returns the name of the test function that called it.
// Only functions with a single *testing.T argument are considered test functions.
func Name(testFn interface{}) string {
	typ := reflect.TypeOf(testFn)
	if typ.Kind() != reflect.Func {
		panic(ErrMsg(reflect.ValueOf(testFn), "not a function"))
	}
	name := runtime.FuncForPC(reflect.ValueOf(testFn).Pointer()).Name()
	name = lastElem(name)
	if typ.NumIn() != 1 || typ.NumOut() > 0 || !strings.HasPrefix(name, "Test") {
		panic(ErrMsg(name, "not a test function"))
	}
	if !typ.In(0).AssignableTo(reflect.TypeOf(&testing.T{})) {
		panic(ErrMsg(name, "test function missing *testing.T argument"))
	}
	return name
}

// IsCaller returns true if the calling function is a test function with the given name.
func IsCaller(testName string) bool {
	// get the call frame of the calling Test function
	testCallFrame := CallFrame()
	// strip the package name from the Test function
	testFnName := stripPkg(testCallFrame.Function)
	// strip the subtest function name from the Test function
	rootTestName := firstElem(testFnName)
	return strings.HasPrefix(testName, rootTestName)
}

// CallerName returns the name of the test function that called it.
func CallerName() string {
	frame := CallFrame()
	return lastElem(frame.Function)
}

func ErrMsg(testFn interface{}, msg string) error {
	frame := CallFrame()
	return fmt.Errorf("%s:%d: %s: %v", filepath.Base(frame.File), frame.Line, msg, testFn)
}

func stripPkg(name string) string {
	start := strings.LastIndex(name, "/") + 1
	dot := strings.Index(name[start:], ".") + 1
	return name[start+dot:]
}

func lastElem(name string) string {
	return name[strings.LastIndex(name, ".")+1:]
}

func firstElem(name string) string {
	end := strings.Index(name, ".")
	if end < 0 {
		// No dots found in function name
		return name
	}
	return name[:end]
}
