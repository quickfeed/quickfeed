package ci

import (
	"bytes"
	"strings"

	"github.com/autograde/quickfeed/kit/score"
)

const truncateMsg = `

... truncated output ...

`

// truncateLog returns the log output truncated at the nearest line before the truncate point.
// The returned log includes score lines found in the middle segment unless the middle segment's size exceeds max.
// The returned log also includes the last segment of size given by last.
func truncateLog(stdout *bytes.Buffer, truncate, last, max int) string {
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
	if len(middleSegment) < max {
		// find score lines in the middle segment that otherwise gets truncated
		scoreLines = findScoreLines(middleSegment)
	}
	return all[0:startMiddleSegment] + scoreLines + truncateMsg + all[startLastSegment:]
}

func findScoreLines(lines string) string {
	scoreLines := make([]string, 0)
	for _, line := range strings.Split(lines, "\n") {
		// check if line has expected JSON score string
		if score.HasPrefix(line) {
			scoreLines = append(scoreLines, line)
		}
	}
	return strings.Join(scoreLines, "\n")
}
