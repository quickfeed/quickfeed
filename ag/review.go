package ag

import (
	"encoding/json"
	"strings"
)

// MarshalReviewString generates a slice of JSON strings to store in the database
func (r *Review) MarshalReviewString() error {
	str := make([]string, 0)
	for _, bm := range r.Benchmarks {
		b, err := json.Marshal(bm)
		if err != nil {
			return err
		}
		str = append(str, string(b))
	}
	r.Review = strings.Join(str, "; ")
	return nil
}

// UnmarshalReviewString converts database string with all submission reviews
// into protobuf messages
func (r *Review) UnmarshalReviewString() error {
	rs := strings.Split(r.Review, ";")
	bms := make([]*GradingBenchmark, 0)
	for _, s := range rs {
		bm := GradingBenchmark{}
		if err := json.Unmarshal([]byte(s), &bm); err != nil {
			return err
		}
		bms = append(bms, &bm)
	}
	r.Benchmarks = bms
	return nil
}

// MakeSubmissionReviews unmarshalls review string for a submission
func (s Submission) MakeSubmissionReviews() error {
	for _, r := range s.Reviews {
		if err := r.UnmarshalReviewString(); err != nil {
			return err
		}
	}
	return nil
}
