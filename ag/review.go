package ag

import (
	fmt "fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
)

// MakeReviewString generates a slice of JSON strings
// to store in the database
func (r *Review) MakeReviewString() error {
	fmt.Println("Marshalling reviews: ", r.Benchmarks)
	m := jsonpb.Marshaler{EnumsAsInts: true}
	str := make([]string, 0)
	for _, bm := range r.Benchmarks {
		s, err := m.MarshalToString(bm)
		if err != nil {
			fmt.Println("Failed to marshall ", bm, " to string: ", err.Error())
			return err
		}
		str = append(str, s)

	}
	fmt.Println("Reviews marshalled successfully: ", strings.Join(str, ";"))
	r.Review = strings.Join(str, "; ")
	return nil
}

// FromReviewString converts database string with all submission reviews
// into protobuf messages
func (r *Review) FromReviewString() error {
	rs := strings.Split(r.Review, ";")
	bms := make([]*GradingBenchmark, 0)
	for _, s := range rs {
		fmt.Println("Unmarshalling review: ", s)
		bm := GradingBenchmark{}
		if err := jsonpb.UnmarshalString(s, &bm); err != nil {
			fmt.Println("Failed to unmarshall ", s, ": ", err.Error())
			// return err
		}
		bms = append(bms, &bm)
	}
	fmt.Println("Unmarshalled successfully: ", bms)
	r.Benchmarks = bms
	return nil
}

// MakeSubmissionReviews unmarshalls review string for a submission
func (s Submission) MakeSubmissionReviews() {
	for _, r := range s.Reviews {
		r.FromReviewString()
		// TODO: design a proper error handling
		// for marshalling/unmarshalling methods
	}
}
