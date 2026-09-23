package ci

import (
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
func capOutput(out string) string {
	if len(out) <= logOutputHead+logOutputTail {
		return out
	}
	head, _, tail := splitOutput(out, logOutputHead, logOutputTail)
	return head + truncateMsg + tail
}

// truncateLog shortens out to its first headLen and last tailLen bytes, as
// capOutput does, but keeps the score lines found in the part left out, so
// that a run printing too much still gets its score. The part left out is
// scanned only if it is shorter than maxScan bytes. Output that already fits
// is returned unchanged.
func truncateLog(out string, headLen, tailLen, maxScan int) string {
	if len(out) <= headLen+tailLen {
		return out
	}
	head, middle, tail := splitOutput(out, headLen, tailLen)
	// TODO(meling) Remove this code when we implement gRPC-based score reporting
	// TODO(meling) Can also remove the maxScan argument in this case
	scoreLines := "too much output data to scan (skipping; fix your code)"
	if len(middle) < maxScan {
		scoreLines = findScoreLines(middle)
	}
	return head + scoreLines + truncateMsg + tail
}

// splitOutput splits out into a head of at most headLen bytes, a tail of at
// most tailLen bytes, and the middle that lies between them. It requires
// len(out) > headLen+tailLen.
//
// Each side is cut back to a line boundary where it has one. A line longer
// than the cut has none, and dropping it would leave nothing at all of a run
// whose output is one long line, which is a shape a test framework or a
// stack trace really does produce. Such a side is cut mid-line instead, at a
// rune boundary: half a line still says what the run printed, while half a
// rune is not a character and would show up as garbled text.
func splitOutput(out string, headLen, tailLen int) (head, middle, tail string) {
	head = out[:headLen]
	if i := strings.LastIndex(head, "\n"); i >= 0 {
		head = head[:i+1]
	} else {
		head = trimPartialRuneSuffix(head)
	}
	tail = out[len(out)-tailLen:]
	// The newline ending the output ends its last line; it is not the start of
	// a line that the tail could begin at.
	if i := strings.Index(strings.TrimSuffix(tail, "\n"), "\n"); i >= 0 {
		tail = tail[i+1:]
	} else {
		tail = trimPartialRunePrefix(tail)
	}
	return head, out[len(head) : len(out)-len(tail)], tail
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
