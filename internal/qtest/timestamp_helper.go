package qtest

import (
	"testing"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Timestamp returns a protobuf timestamp representation of the given string time.
func Timestamp(t *testing.T, tim string) *timestamppb.Timestamp {
	t.Helper()
	timeTime, err := time.Parse(qf.TimeLayout, tim)
	if err != nil {
		t.Fatal(err)
	}
	return timestamppb.New(timeTime)
}
