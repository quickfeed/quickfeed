package ci

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
)

func TestLogTruncate(t *testing.T) {
	const (
		logLines  = "want \n only this \n part of the \n output \n but not this part \n because it is \n too long \n but the last part \n we do want"
		scoreLine = `{"Secret":"For Your Eyes Only","TestName":"JamesBond","Score":100,"MaxScore":100,"Weight":1}`
	)
	withScoreLine := logLines[0:77] + scoreLine + "\n" + logLines[77:]
	tests := []struct {
		name     string
		truncate int
		last     int
		max      int
		in       string
		want     string
	}{
		{
			name:     "output that fits is left alone",
			truncate: 100, last: 19, max: 1000,
			in:   logLines,
			want: logLines,
		},
		{
			name:     "sides without a line boundary are cut mid-line",
			truncate: 4, last: 5, max: 1000,
			in:   logLines,
			want: "want" + truncateMsg + " want",
		},
		{
			name:     "head keeps its whole lines",
			truncate: 6, last: 5, max: 1000,
			in:   logLines,
			want: "want \n" + truncateMsg + " want",
		},
		{
			name:     "head is cut back to its last line boundary",
			truncate: 45, last: 5, max: 1000,
			in:   logLines,
			want: "want \n only this \n part of the \n output \n" + truncateMsg + " want",
		},
		{
			name:     "tail is cut forward to its first line boundary",
			truncate: 45, last: 21, max: 1000,
			in:   logLines,
			want: "want \n only this \n part of the \n output \n" + truncateMsg + " we do want",
		},
		{
			name:     "tail keeps its whole lines",
			truncate: 45, last: 32, max: 1000,
			in:   logLines,
			want: "want \n only this \n part of the \n output \n" + truncateMsg + " but the last part \n we do want",
		},
		{
			name:     "score lines left out are kept",
			truncate: 45, last: 32, max: 1000,
			in:   withScoreLine,
			want: "want \n only this \n part of the \n output \n" + scoreLine + truncateMsg + " but the last part \n we do want",
		},
		{
			name:     "too much left out to scan for score lines",
			truncate: 45, last: 32, max: 10,
			in:   withScoreLine,
			want: "want \n only this \n part of the \n output \n" + "too much output data to scan (skipping; fix your code)" + truncateMsg + " but the last part \n we do want",
		},
		{
			// Output with no line boundary at all used to be kept whole.
			name:     "a single long line is still truncated",
			truncate: 10, last: 10, max: 1000,
			in:   strings.Repeat("x", 100),
			want: strings.Repeat("x", 10) + truncateMsg + strings.Repeat("x", 10),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := truncateLog(test.in, test.truncate, test.last, test.max)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("truncateLog() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCapOutput(t *testing.T) {
	line := strings.Repeat("x", 99) + "\n" // 100 bytes
	long := strings.Repeat(line, 100)      // 10,000 bytes, well over the cap
	tests := []struct {
		name string
		out  string
		want func(string) error
	}{
		{
			name: "short output is left alone",
			out:  "all tests passed\n",
			want: func(got string) error {
				if got != "all tests passed\n" {
					return fmt.Errorf("capOutput() = %q, want the output unchanged", got)
				}
				return nil
			},
		},
		{
			name: "output at the cap is left alone",
			out:  strings.Repeat("y", logOutputHead+logOutputTail),
			want: func(got string) error {
				if len(got) != logOutputHead+logOutputTail {
					return fmt.Errorf("len(capOutput()) = %d, want %d", len(got), logOutputHead+logOutputTail)
				}
				return nil
			},
		},
		{
			name: "long output keeps its head and its tail",
			out:  "FIRST LINE\n" + long + "LAST LINE",
			want: func(got string) error {
				if !strings.HasPrefix(got, "FIRST LINE\n") {
					return fmt.Errorf("capOutput() = %q..., want it to start with the first line", got[:20])
				}
				if !strings.HasSuffix(got, "LAST LINE") {
					return fmt.Errorf("capOutput() = ...%q, want it to end with the last line", got[len(got)-20:])
				}
				if !strings.Contains(got, truncateMsg) {
					return fmt.Errorf("capOutput() = %q, want it to mark what was left out", got)
				}
				if len(got) > logOutputHead+logOutputTail+len(truncateMsg) {
					return fmt.Errorf("len(capOutput()) = %d, want at most the cap plus the marker", len(got))
				}
				return nil
			},
		},
		{
			name: "cuts on line boundaries",
			out:  long,
			want: func(got string) error {
				head, tail, ok := strings.Cut(got, truncateMsg)
				if !ok {
					return fmt.Errorf("capOutput() = %q, want it to mark what was left out", got)
				}
				// A cut mid-line would leave a partial line of x's on either side.
				for _, part := range []string{head, tail} {
					for l := range strings.SplitSeq(strings.Trim(part, "\n"), "\n") {
						if len(l) != 99 {
							return fmt.Errorf("capOutput() kept a partial line %q of %d bytes, want whole lines", l, len(l))
						}
					}
				}
				return nil
			},
		},
		{
			// The newline that ends the output is not a line boundary for the
			// tail to begin at, or a long last line would leave no tail at all.
			name: "a long last line ending in a newline keeps a tail",
			out:  "FIRST LINE\n" + strings.Repeat("z", 5000) + "\n",
			want: func(got string) error {
				_, tail, ok := strings.Cut(got, truncateMsg)
				if !ok {
					return fmt.Errorf("capOutput() = %q, want it to mark what was left out", got[:40])
				}
				if want := strings.Repeat("z", logOutputTail-1) + "\n"; tail != want {
					return fmt.Errorf("capOutput() kept a tail of %d bytes, want %d", len(tail), len(want))
				}
				return nil
			},
		},
		{
			// A run whose output is one long line has no line boundary to cut
			// back to. Dropping that side would leave nothing of the run at
			// all, so it is cut mid-line and must at least stay valid UTF-8:
			// the multi-byte runes here straddle both cuts.
			name: "a single long line is cut mid-line, but not mid-rune",
			out:  strings.Repeat("påske—🎄", 2000),
			want: func(got string) error {
				head, tail, ok := strings.Cut(got, truncateMsg)
				if !ok {
					return fmt.Errorf("capOutput() = %q, want it to mark what was left out", got[:40])
				}
				if head == "" || tail == "" {
					return fmt.Errorf("capOutput() kept head %d and tail %d bytes, want a look at both ends", len(head), len(tail))
				}
				if !utf8.ValidString(head) {
					return fmt.Errorf("capOutput() head ends mid-rune: %q", head[len(head)-8:])
				}
				if !utf8.ValidString(tail) {
					return fmt.Errorf("capOutput() tail starts mid-rune: %q", tail[:8])
				}
				if len(got) > logOutputHead+logOutputTail+len(truncateMsg) {
					return fmt.Errorf("len(capOutput()) = %d, want at most the cap plus the marker", len(got))
				}
				return nil
			},
		},
		{
			// Output that was never UTF-8 is not made shorter than the cut on
			// the strength of a rune boundary that was never there.
			name: "output that is not text keeps all but the cut rune",
			out:  strings.Repeat("\xff\xfe", 3000),
			want: func(got string) error {
				head, tail, ok := strings.Cut(got, truncateMsg)
				if !ok {
					return fmt.Errorf("capOutput() = %q, want it to mark what was left out", got[:40])
				}
				if len(head) < logOutputHead-utf8.UTFMax || len(tail) < logOutputTail-utf8.UTFMax {
					return fmt.Errorf("capOutput() kept head %d and tail %d bytes, want close to %d each", len(head), len(tail), logOutputHead)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.want(capOutput(test.out)); err != nil {
				t.Error(err)
			}
		})
	}
}
