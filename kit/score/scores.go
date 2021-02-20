package score

func NewScores() *Scores {
	return &Scores{
		TestNames: make([]string, 0),
		ScoreMap:  make(map[string]*Score),
	}
}

// AddScore adds the given score to the set of scores.
// This method assumes that the provided score object is valid.
func (s *Scores) AddScore(sc *Score) {
	testName := sc.GetTestName()
	if current, found := s.ScoreMap[testName]; found {
		if current.GetScore() != 0 {
			// We reach here only if a second non-zero score is found
			// Mark it as faulty with -1.
			sc.Score = -1
		}
	} else {
		// New test: record in TestNames
		s.TestNames = append(s.TestNames, testName)
	}

	// Record score object if:
	// - current score is nil or zero, or
	// - the first score was zero.
	s.ScoreMap[testName] = sc
}

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func (s *Scores) Validate() error {
	for _, sc := range s.GetScoreMap() {
		if err := sc.IsValid(hiddenSecret); err != nil {
			return err
		}
	}
	return nil
}

// Sum returns the total score computed over the set of recorded scores.
// The total is a grade in the range 0-100.
// This method must only be called after Validate has returned nil.
func (s *Scores) Sum() uint32 {
	totalWeight := float32(0)
	var max, score, weight []float32
	for _, ts := range s.GetScoreMap() {
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
