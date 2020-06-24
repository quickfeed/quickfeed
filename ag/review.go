package ag

import (
	"strings"

	"github.com/golang/protobuf/jsonpb"
)

// MakeReviewString generates a slice of JSON strings
// to store in the database
func (r *Review) MakeReviewString() error {
	m := jsonpb.Marshaler{EnumsAsInts: true}
	str := make([]string, 0)
	for _, bm := range r.Benchmarks {
		s, err := m.MarshalToString(bm)
		if err != nil {
			return err
		}
		str = append(str, s)

	}
	r.Review = strings.Join(str, "; ")
	return nil
}

// FromReviewString converts database string with all submission reviews
// into protobuf messages
func (r *Review) FromReviewString() error {
	rs := strings.Split(r.Review, ";")
	bms := make([]*GradingBenchmark, 0)
	for _, s := range rs {
		bm := GradingBenchmark{}
		if err := jsonpb.UnmarshalString(s, &bm); err != nil {
			return err
		}
		bms = append(bms, &bm)
	}
	r.Benchmarks = bms
	return nil
}

// MakeSubmissionReviews unmarshalls review string for a submission
func (s Submission) MakeSubmissionReviews() {
	for _, r := range s.Reviews {
		r.FromReviewString()
	}
}
