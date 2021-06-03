package score

import (
	"encoding/json"
)

// TODO(meling) Replace ci.Result with score.Result - should remove method
// TODO(meling) Should remove method and use buildInfo and scores directly

// Marshal returns marshalled information from the result struct.
func (r *Result) Marshal() (buildInfo string, scores string, err error) {
	bi, e := json.Marshal(r.BuildInfo)
	if e == nil {
		scs, e := json.Marshal(r.Scores)
		if e == nil {
			buildInfo = string(bi)
			scores = string(scs)
		}
	}
	err = e
	return
}

// TODO(meling) Replace ci.Result with score.Result - should remove method??

// TotalScore returns the total score for this execution result.
func (r *Result) TotalScore() uint32 {
	// r.Scores
	// return Sum
	return Total(r.Scores) // Total method is deprecated; use scores.Sum()
}
