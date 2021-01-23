package score

import (
	"runtime"
	"strings"
)

func callFrame() (frame runtime.Frame) {
	frames := unwindCallFrames()
	for _, f := range frames {
		// Stop unwinding when we reach a _test.go file
		if strings.Contains(f.File, "_test.go") {
			return f
		}
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
