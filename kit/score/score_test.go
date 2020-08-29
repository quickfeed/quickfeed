package score

import (
	"os"
	"testing"
)

var theSecret = "my secret code"

func init() {
	GlobalSecret = theSecret
}

func fibonacci(n uint) uint {
	if n <= 1 || n == 5 {
		return n
	}
	if n < 5 {
		return n - 1
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

var fibonacciTests = []struct {
	in, want uint
}{
	{0, 0},
	{1, 1},
	{2, 1},
	{3, 2},
	{4, 3},
	{5, 5},
	{6, 8},
	{7, 13},
	{8, 21},
	{9, 34},
	{10, 155},
	{12, 154},
	{16, 1987},
	{20, 26765},
}

func TestFibonacci(t *testing.T) {
	sc := NewScoreMax(t, len(fibonacciTests)*2, 20)
	defer sc.WriteJSON(os.Stdout)

	for _, ft := range fibonacciTests {
		out := fibonacci(ft.in)
		if out != ft.want {
			sc.Dec()
		}
	}
	if sc.Score != 24 {
		t.Errorf("expected 24 tests to pass, but got %v", sc.Score)
	}
	if sc.TestName != "TestFibonacci" {
		t.Errorf("expected TestName=TestFibonacci, but got %v", sc.TestName)
	}
}

var nonJSONLog = []string{
	"here is some output",
	"some other output",
	"line contains " + theSecret,
	theSecret + " should not be revealed",
}

func TestParseNonJSONStrings(t *testing.T) {
	for _, s := range nonJSONLog {
		sc, err := Parse(s, GlobalSecret)
		if err == nil {
			t.Errorf("Expected '%v', got '<nil>'", ErrScoreNotFound.Error())
		}
		if sc != nil {
			t.Errorf("Got unexpected score object '%v', wanted '<nil>'", sc)
		}
	}
}

var jsonLog = []struct {
	in          string
	max, weight int
	err         error
}{
	{`{"Secret":"` + theSecret + `","TestName":"TestParseJSONStrings","Score":0,"MaxScore":10,"Weight":10}`,
		10, 10,
		nil,
	},
	{`{"Secret":"the wrong secret","TestName":"TestParseJSONStrings","Score":0,"MaxScore":10,"Weight":10}`,
		-1, -1,
		ErrScoreNotFound,
	},
}

// Equal returns true if sc equals other. Ignores the Secret field.
func (sc *Score) Equal(other *Score) bool {
	return other != nil &&
		sc.TestName == other.TestName &&
		sc.Score == other.Score &&
		sc.MaxScore == other.MaxScore &&
		sc.Weight == other.Weight
}

func TestParseJSONStrings(t *testing.T) {
	for _, s := range jsonLog {
		sc, err := Parse(s.in, GlobalSecret)
		var expectedScore *Score
		if s.max > 0 {
			expectedScore = NewScore(t, s.max, s.weight)
		}
		if sc != expectedScore || err != s.err {
			if !expectedScore.Equal(sc) || err != s.err {
				t.Errorf("Failed to parse:\n%v\nGot: '%v', '%v'\nExp: '%v', '%v'",
					s.in, sc, err, expectedScore, s.err)
			}
			if sc != nil && sc.Secret == GlobalSecret {
				t.Errorf("Parse function failed to hide global secret: %v", sc.Secret)
			}
		}
	}
}
