package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGetCreateDeleteTokenRecords(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	tokens := []*pb.UpdateTokenRecord{
		{
			UserID: 1,
		},
		{
			UserID: 2,
		},
		{
			UserID: 3,
		},
	}
	for _, token := range tokens {
		if err := db.CreateTokenRecord(token); err != nil {
			t.Fatal(err)
		}
	}
	savedTokens, err := db.GetTokenRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != len(savedTokens) {
		t.Errorf("have %d tokens, saved %d tokens", len(tokens), len(savedTokens))
	}
	if diff := cmp.Diff(savedTokens, tokens, protocmp.Transform()); diff != "" {
		t.Errorf("incorrect token records were saved (-have +want):\n%s", diff)
	}

	// remove token with ID 2
	tokenToRemove := savedTokens[1]
	wantTokens := []*pb.UpdateTokenRecord{savedTokens[0], savedTokens[2]}
	if err := db.DeleteTokenRecord(tokenToRemove); err != nil {
		t.Fatal(err)
	}
	updatedTokens, err := db.GetTokenRecords()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTokens, updatedTokens, protocmp.Transform()); diff != "" {
		t.Errorf("incorrect token records after delete (-have +want):\n%s", diff)
	}
}
