package score

import (
	"testing"
)

func TestCallFrame(t *testing.T) {
	frame := callFrame()
	t.Logf("- %s:%d %s\n", frame.File, frame.Line, frame.Function)
}

func TestFrame(t *testing.T) {
	frames := unwindCallFrames()
	for _, frame := range frames {
		t.Logf("- %s:%d %s\n", frame.File, frame.Line, frame.Function)
	}
}

func TestFrame2(t *testing.T) {
	t.Run("SubTest2", func(t *testing.T) {
		frames := unwindCallFrames()
		for _, frame := range frames {
			t.Logf("- %s:%d %s\n", frame.File, frame.Line, frame.Function)
		}
	})
}

func TestFrame3(t *testing.T) {
	t.Run("SubTest3", func(t *testing.T) {
		t.Run("SubSubTest1", func(t *testing.T) {
			frames := unwindCallFrames()
			for _, frame := range frames {
				t.Logf("- %s:%d %s\n", frame.File, frame.Line, frame.Function)
			}
		})
		t.Run("SubSubTest2", func(t *testing.T) {
			frames := unwindCallFrames()
			for _, frame := range frames {
				t.Logf("- %s:%d %s\n", frame.File, frame.Line, frame.Function)
			}
		})
		t.Run("SubSubTest3", func(t *testing.T) {
			frames := unwindCallFrames()
			for _, frame := range frames {
				t.Logf("- %s:%d %s\n", frame.File, frame.Line, frame.Function)
			}
		})
	})
}
