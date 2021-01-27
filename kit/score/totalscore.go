package score

const (
	errScoreInterval = "Score must be in the interval [0, MaxScore]"
	errMaxScore      = "MaxScore must be greater than 0"
	errWeight        = "Weight must be greater than 0"
)

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func Validate() error {
	for testName, ts := range scores {
		if ts.MaxScore <= 0 {
			return errMsg(testName, errMaxScore)
		}
		if ts.Weight <= 0 {
			return errMsg(testName, errWeight)
		}
		if ts.Score < 0 || ts.Score > ts.MaxScore {
			return errMsg(testName, errScoreInterval)
		}
	}
	return nil
}

// Sum returns the total score computed over the set of recorded scores.
// The total is a grade in the range 0-100.
// This method must only be called after Validate has returned nil.
func Sum() uint32 {
	totalWeight := float32(0)
	var max, score, weight []float32
	for _, ts := range scores {
		totalWeight += float32(ts.Weight)
		weight = append(weight, float32(ts.Weight))
		score = append(score, float32(ts.Score))
		max = append(max, float32(ts.MaxScore))
	}
	total := float32(0)
	for i := 0; i < len(score); i++ {
		if score[i] > max[i] {
			score[i] = max[i]
		}
		total += ((score[i] / max[i]) * (weight[i] / totalWeight))
	}
	return uint32(total * 100)
}
