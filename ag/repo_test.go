package ag_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
)

func TestGetTestURL(t *testing.T) {
	want := "https://github.com/dat320-2020/" + pb.TestsRepo
	repo := &pb.Repository{
		HTMLURL: "https://github.com/dat320-2020/meling-labs",
	}
	got := repo.GetTestURL()
	if got != want {
		t.Errorf("GetTestURL() = %s, want %s", got, want)
	}
}
