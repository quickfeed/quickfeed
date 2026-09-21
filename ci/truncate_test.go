package ci

import (
	"strings"
	"testing"

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
			// The newline ending the output is not a line boundary for the
			// tail to begin at, or a long last line would leave no tail.
			name:     "a long last line ending in a newline keeps a tail",
			truncate: 10, last: 6, max: 1000,
			in:   "first\n" + strings.Repeat("z", 50) + "\n",
			want: "first\n" + truncateMsg + "zzzzz\n",
		},
		{
			// The multi-byte runes straddle both cuts.
			name:     "a single long line is not cut mid-rune",
			truncate: 10, last: 10, max: 1000,
			in:   strings.Repeat("å", 50),
			want: strings.Repeat("å", 5) + truncateMsg + strings.Repeat("å", 5),
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

func TestLogTruncateKeepsScoresInEveryForm(t *testing.T) {
	// A score survives truncation so that the run is still scored. Since Go
	// 1.25 a test may report it through testing.T.Attr, which prints it as an
	// "=== ATTR" line, and as an event when the run script uses go test -json.
	const (
		filler    = "want \n only this \n part of the \n output \n but not this part \n because it is \n too long \n but the last part \n we do want"
		score     = `{"Secret":"For Your Eyes Only","TestName":"JamesBond","Score":100,"MaxScore":100,"Weight":1}`
		attrLine  = `=== ATTR  JamesBond score ` + score
		attrEvent = `{"Time":"2026-09-21T22:45:05.789035+02:00","Action":"attr","Package":"lab1","Test":"JamesBond","Key":"score","Value":"{\"Secret\":\"For Your Eyes Only\",\"TestName\":\"JamesBond\",\"Score\":100,\"MaxScore\":100,\"Weight\":1}"}`
		// A course that prints the score itself, under a go test -json run.
		outputEvent = `{"Time":"2026-09-21T22:45:05.789063+02:00","Action":"output","Package":"lab1","Test":"JamesBond","Output":"{\"Secret\":\"For Your Eyes Only\",\"TestName\":\"JamesBond\",\"Score\":100,\"MaxScore\":100,\"Weight\":1}\n"}`
	)
	for _, carrier := range []string{score, attrLine, attrEvent, outputEvent} {
		t.Run(strings.Fields(carrier)[0], func(t *testing.T) {
			got := truncateLog(filler[0:77]+carrier+"\n"+filler[77:], 45, 15, 1000)
			if !strings.Contains(got, carrier) {
				t.Errorf("truncateLog() = %q, want it to keep %q", got, carrier)
			}
		})
	}
}
