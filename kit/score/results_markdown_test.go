package score_test

import (
	"testing"

	"github.com/quickfeed/quickfeed/kit/score"
)

func TestMarkdownComment(t *testing.T) {
	expected := `## Test results from latest push

| Test Name | Score | Weight | % of Total |
| :-------- | ----: | -----: | ---------: |
| Test1 | 5/7 | 2 | 11.9% |
| Test2 | 3/9 | 3 | 8.3% |
| Test3 | 8/8 | 5 | 41.7% |
| Test4 | 2/5 | 1 | 3.3% |
| Test5 | 5/7 | 1 | 6.0% |
| **Total** | | | **71.2%** |

Reviewers are assigned once the total score reaches 80%.
`
	results := &score.Results{
		Scores: []*score.Score{
			{TestName: "Test1", TaskName: "1", Score: 5, MaxScore: 7, Weight: 2},
			{TestName: "Test2", TaskName: "1", Score: 3, MaxScore: 9, Weight: 3},
			{TestName: "Test3", TaskName: "1", Score: 8, MaxScore: 8, Weight: 5},
			{TestName: "Test4", TaskName: "1", Score: 2, MaxScore: 5, Weight: 1},
			{TestName: "Test5", TaskName: "1", Score: 5, MaxScore: 7, Weight: 1},
			{TestName: "Test6", TaskName: "2", Score: 5, MaxScore: 7, Weight: 1},
			{TestName: "Test7", TaskName: "3", Score: 5, MaxScore: 7, Weight: 1},
		},
	}
	body := results.MarkdownComment("1", 80)
	if body != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, body)
	}
}
