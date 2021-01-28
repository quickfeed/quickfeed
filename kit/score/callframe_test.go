package score

import (
	"path/filepath"
	"testing"
)

func TestCallFrame(t *testing.T) {
	frame := callFrame()
	expectedFunc := pkg + t.Name()
	if frame.Function != expectedFunc {
		t.Errorf("callFrame().Function = %s, expected %s", frame.Function, expectedFunc)
	}
	if filepath.Base(frame.File) != "callframe_test.go" {
		t.Errorf("callFrame().File = %s, expected %s", filepath.Base(frame.File), "callframe_test.go")
	}
	expectedLine := 9
	if frame.Line != expectedLine {
		t.Errorf("callFrame().Line = %d, expected %d", frame.Line, expectedLine)
	}
}

func TestFrame(t *testing.T) {
	frames := unwindCallFrames()
	if len(frames) != 1 {
		t.Errorf("len(frames)=%d, expected 1", len(frames))
	}
	expectedFunc := pkg + t.Name()
	if frames[0].Function != expectedFunc {
		t.Errorf("unwindCallFrames().Function = %s, expected %s", frames[0].Function, expectedFunc)
	}
}

func TestFrame2(t *testing.T) {
	mainTest := t.Name()
	t.Run("SubTest", func(t *testing.T) {
		frames := unwindCallFrames()
		if len(frames) != 1 {
			t.Errorf("len(frames)=%d, expected 1", len(frames))
		}
		expectedFunc := pkg + mainTest + ".func1"
		if frames[0].Function != expectedFunc {
			t.Errorf("unwindCallFrames().Function = %s, expected %s", frames[0].Function, expectedFunc)
		}
	})
}

func TestFrame3(t *testing.T) {
	mainTest := t.Name()
	t.Run("SubTest", func(t *testing.T) {
		t.Run("SubSubTest1", func(t *testing.T) {
			frames := unwindCallFrames()
			if len(frames) != 1 {
				t.Errorf("len(frames)=%d, expected 1", len(frames))
			}
			expectedFunc := pkg + mainTest + ".func1.1"
			if frames[0].Function != expectedFunc {
				t.Errorf("unwindCallFrames().Function = %s, expected %s", frames[0].Function, expectedFunc)
			}
		})
		t.Run("SubSubTest2", func(t *testing.T) {
			frames := unwindCallFrames()
			if len(frames) != 1 {
				t.Errorf("len(frames)=%d, expected 1", len(frames))
			}
			expectedFunc := pkg + mainTest + ".func1.2"
			if frames[0].Function != expectedFunc {
				t.Errorf("unwindCallFrames().Function = %s, expected %s", frames[0].Function, expectedFunc)
			}
		})
		t.Run("SubSubTest3", func(t *testing.T) {
			frames := unwindCallFrames()
			if len(frames) != 1 {
				t.Errorf("len(frames)=%d, expected 1", len(frames))
			}
			expectedFunc := pkg + mainTest + ".func1.3"
			if frames[0].Function != expectedFunc {
				t.Errorf("unwindCallFrames().Function = %s, expected %s", frames[0].Function, expectedFunc)
			}
		})
	})
}
