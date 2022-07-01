package score

import "fmt"

// MarkdownComment returns a markdown formatted feedback comment of the test results.
// Only the test scores associated with the supplied task are included in the table.
// An example table is shown below.
//
//  ## Test results from latest push
//
//	| Test Name | Score | Weight | % of Total |
//	| :-------- | ----: | -----: | ---------: |
//  | Test 1    |   2/4 |      1 |       6.3% |
//  | Test 2    |   1/4 |      2 |       6.3% |
//  | Test 3    |   3/4 |      5 |      46.9% |
//  | Total     |       |        |      59.5% |
//
// 	Reviewers are assigned once the total score reaches 80%.
//
func (r *Results) MarkdownComment(taskLocalName string, scoreLimit uint32) string {
	body := `## Test results from latest push

| Test Name | Score | Weight | % of Total |
| :-------- | ----: | -----: | ---------: |
`

	total, totalWeight := r.internalSum(taskLocalName)
	for _, sc := range r.Scores {
		if sc.TaskName != taskLocalName {
			continue
		}
		weightedScore := sc.weightedScore(totalWeight)
		body += fmt.Sprintf("| %s | %d/%d | %d | %.1f%% |\n",
			sc.TestName, sc.Score, sc.MaxScore, sc.Weight, weightedScore*100)
	}
	body += fmt.Sprintf("| **Total** | | | **%.1f%%** |\n\n", total*100)
	body += fmt.Sprintf("Reviewers are assigned once the total score reaches %d%%.\n", scoreLimit)
	return body
}
