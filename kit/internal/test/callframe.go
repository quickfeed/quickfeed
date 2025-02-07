package test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	testPkg  = "github.com/quickfeed/quickfeed/kit/internal/test."
	scorePkg = "github.com/quickfeed/quickfeed/kit/score."
)

// CallFrame returns the call frame of the Test function that
// called one of the registry functions.
func CallFrame() (frame runtime.Frame) {
	frames := unwindCallFrames()
	for _, f := range frames {
		// The call frame must be in a _test.go file
		if strings.HasSuffix(f.File, "_test.go") {
			return f
		}
		// Special case handling for TestCallFrame
		if strings.HasPrefix(f.Function, testPkg+"TestCallFrame") {
			return f
		}
		// Ignore functions in kit/score and kit/internal/test packages
		if strings.HasPrefix(f.Function, testPkg) || strings.HasPrefix(f.Function, scorePkg) {
			continue
		}
		// Only Test functions can call the CallFrame function
		panic(fmt.Errorf("%s:%d: %s: %s", filepath.Base(f.File), f.Line, stripPkg(f.Function), "unauthorized lookup"))
	}
	return
}

func unwindCallFrames() []runtime.Frame {
	// Ask runtime.Callers for up to 10 pcs, excluding runtime.Callers and this function itself.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	foundFrames := make([]runtime.Frame, 0)
	for {
		frame, more := frames.Next()
		// Stop unwinding when we reach package testing.
		if strings.Contains(frame.File, "testing/") {
			break
		}
		foundFrames = append(foundFrames, frame)
		if !more {
			break
		}
	}
	return foundFrames
}
