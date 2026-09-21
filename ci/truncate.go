package ci

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/quickfeed/quickfeed/kit/score"
)

const truncateMsg = `

... truncated output ...

`

// logOutputHead and logOutputTail bound the output a successful run copies to
// the course log. A failure logs its whole output, since that is what the
// teacher is there to read; a success is logged so a teacher can see what a
// passing run produced, which the head and the tail answer without keeping a
// full run's output for every student for the log's whole retention.
const (
	logOutputHead = 2048 // bytes
	logOutputTail = 2048 // bytes
)

// capOutput shortens out to its first logOutputHead and last logOutputTail
// bytes, with truncateMsg marking what was left out. Output that already fits
// is returned unchanged.
//
// Each side is cut back to a line boundary where it has one. A line longer
// than the cap has none, and dropping it would leave nothing at all of a run
// whose output is one long line, which is a shape a test framework or a
// stack trace really does produce. Such a side is cut mid-line instead, at a
// rune boundary: half a line still says what the run printed, while half a
// rune is only mojibake.
func capOutput(out string) string {
	if len(out) <= logOutputHead+logOutputTail {
		return out
	}
	head := out[:logOutputHead]
	if i := strings.LastIndex(head, "\n"); i >= 0 {
		head = head[:i]
	} else {
		head = trimPartialRuneSuffix(head)
	}
	tail := out[len(out)-logOutputTail:]
	if i := strings.Index(tail, "\n"); i >= 0 {
		tail = tail[i+1:]
	} else {
		tail = trimPartialRunePrefix(tail)
	}
	return head + truncateMsg + tail
}

// trimPartialRuneSuffix drops the incomplete encoding a cut at an arbitrary
// byte can leave at the end of s. One rune is at most utf8.UTFMax bytes, so
// the trimming is bounded there: output that was never UTF-8 to begin with,
// such as a run that printed binary, is left all but whole.
func trimPartialRuneSuffix(s string) string {
	for range utf8.UTFMax - 1 {
		if r, size := utf8.DecodeLastRuneInString(s); r != utf8.RuneError || size > 1 {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}

// trimPartialRunePrefix is trimPartialRuneSuffix for the start of s, where the
// cut leaves the continuation bytes of a rune whose first byte was left behind.
func trimPartialRunePrefix(s string) string {
	for range utf8.UTFMax - 1 {
		if r, size := utf8.DecodeRuneInString(s); r != utf8.RuneError || size > 1 {
			break
		}
		s = s[1:]
	}
	return s
}

// truncateLog returns the log output truncated at the nearest line before the truncate point.
// The returned log includes score lines found in the middle segment unless the middle segment's
// size exceeds maxLen. The returned log also includes the last segment of size given by last.
func truncateLog(stdout *bytes.Buffer, truncate, last, maxLen int) string {
	// converting to string here;
	// could be done more efficiently using stdout.Truncate(maxLogSize)
	// but then we wouldn't get the last part
	all := stdout.String()
	// find the last full line to keep before the truncate point
	startMiddleSegment := strings.LastIndex(all[0:truncate], "\n") + 1
	// find the last full line to truncate and scan for score lines, before the last segment to output
	startLastSegment := strings.LastIndex(all[0:len(all)-last], "\n") + 1

	// TODO(meling) Remove this code when we implement gRPC-based score reporting
	// TODO(meling) Can also remove the max argument in this case
	middleSegment := all[startMiddleSegment:startLastSegment]
	// score lines will normally replace this string, unless too much output
	scoreLines := "too much output data to scan (skipping; fix your code)"
	// scan middle segment for score lines only if middle segment is less than max
	if len(middleSegment) < maxLen {
		// find score lines in the middle segment that otherwise gets truncated
		scoreLines = findScoreLines(middleSegment)
	}
	return all[0:startMiddleSegment] + scoreLines + truncateMsg + all[startLastSegment:]
}

func findScoreLines(lines string) string {
	scoreLines := make([]string, 0)
	for line := range strings.SplitSeq(lines, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			scoreLines = append(scoreLines, line)
		}
	}
	return strings.Join(scoreLines, "\n")
}
