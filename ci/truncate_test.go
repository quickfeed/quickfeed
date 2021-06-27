package ci

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

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
