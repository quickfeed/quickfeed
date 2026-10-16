package wrapper

import (
	"sync"
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

var scores = score.NewRegistry()

func init() {
	scores.Add(TestTopError, 2, 1)
	scores.Add(TestError, 2, 1)
	scores.Add(TestFatal, 2, 1)
	scores.Add(TestFatalf, 2, 1)
	scores.Add(TestSubError, 2, 1)
	scores.Add(TestParallel, 4, 1)
	scores.Add(TestTopPanic, 2, 1)
	scores.Add(TestSubPanic, 2, 1)
	scores.Add(TestAfterPanic, 2, 1)
	scores.Add(TestConcurrent, 40, 1)
	scores.Add(TestArithmetic, 5, 1)
	scores.Add(TestFail, 2, 1)
}

func TestTopError(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	sc.Errorf("top failure") // top call site
}

func TestError(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	sc.Error("plain", "failure")
}

func TestFatal(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	sc.Fatal("fatal", "failure")
}

func TestFatalf(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	sc.Fatalf("fatal %s", "formatted")
}

func TestSubError(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	sc.Run("child", func(t *score.T) {
		t.Errorf("sub failure") // sub call site
	})
}

func TestParallel(t *testing.T) {
	sc := scores.Min(t)
	defer sc.Recover()
	for i := range 4 {
		sc.Run(string(rune('a'+i)), func(t *score.T) {
			t.Parallel()
			t.Inc()
			t.Logf("parallel done %s", t.Name())
		})
	}
}

func TestTopPanic(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	panic("top panic marker")
}

func TestSubPanic(t *testing.T) {
	sc := scores.Max(t)
	defer sc.Recover()
	sc.Run("child", func(t *score.T) {
		panic("sub panic marker")
	})
}

func TestAfterPanic(t *testing.T) {
	sc := scores.Min(t)
	defer sc.Recover()
	sc.Inc()
}

func TestConcurrent(t *testing.T) {
	sc := scores.Min(t)
	defer sc.Recover()
	var wg sync.WaitGroup
	for range 40 {
		wg.Go(func() {
			sc.Inc()
			sc.Errorf("concurrent failure")
		})
	}
	wg.Wait()
}

func TestArithmetic(t *testing.T) {
	sc := scores.Min(t)
	defer sc.Recover()
	sc.IncBy(4)
	sc.Dec()
	sc.IncBy(2)
	sc.DecBy(1)
}

func TestFail(t *testing.T) {
	sc := scores.Min(t)
	defer sc.Recover()
	sc.Inc()
	sc.Fail()
}

func TestMuted(t *testing.T) {
	sc := score.Muted(t)
	sc.Inc()
	sc.Dec()
	sc.Run("child", func(t *score.T) {
		t.Inc()
	})
}

func TestUnregistered(t *testing.T) {
	scores.Max(t)
}

func TestUnregisteredMin(t *testing.T) {
	scores.Min(t)
}
