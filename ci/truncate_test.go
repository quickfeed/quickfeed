package ci

import (
	"bytes"
	"fmt"
	"io"
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
	tests := []struct {
		truncate int
		last     int
		max      int
		in       string
		want     string
	}{
		{
			truncate: 4, last: 5, max: 1000,
			in:   logLines,
			want: truncateMsg + " we do want",
		},
		{
			truncate: 6, last: 5, max: 1000,
			in:   logLines,
			want: "want \n" + truncateMsg + " we do want",
		},
		{
			truncate: 43, last: 5, max: 1000,
			in:   logLines,
			want: "want \n only this \n part of the \n output \n" + truncateMsg + " we do want",
		},
		{
			truncate: 45, last: 5, max: 1000,
			in:   logLines,
			want: "want \n only this \n part of the \n output \n" + truncateMsg + " we do want",
		},
		{
			truncate: 45, last: 15, max: 1000,
			in:   logLines,
			want: "want \n only this \n part of the \n output \n" + truncateMsg + " but the last part \n we do want",
		},
		{
			truncate: 45, last: 15, max: 1000,
			in:   logLines[0:77] + scoreLine + "\n" + logLines[77:],
			want: "want \n only this \n part of the \n output \n" + scoreLine + truncateMsg + " but the last part \n we do want",
		},
	}
	for _, test := range tests {
		logReader := strings.NewReader(test.in)
		var stdout bytes.Buffer
		_, err := io.Copy(&stdout, logReader)
		if err != nil {
			t.Fatal(err)
		}
		got := truncateLog(&stdout, test.truncate, test.last, test.max)
		if diff := cmp.Diff(test.want, got); diff != "" {
			fmt.Println(got)
			t.Errorf("truncateLog() mismatch (-want +got):\n%s", diff)
		}
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
