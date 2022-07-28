package database_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestGormCreateApplication(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	app, err := db.CreateApplication(&qf.Application{UserID: 10, Name: "test", Description: "test app"})
	if err != nil || len(app.GetSecret()) == 0 {
		t.Errorf("have error '%v' have key '%v'", err, app.GetSecret())
	}

	gotApp, err := db.GetApplication(10, app.GetClientID(), app.GetSecret())
	if err != nil {
		t.Errorf("have error '%v'", err)
	}
	// If GetApplication does not return an error, it implies that the provided secret is correct.

	app.ID = 0
	app.Secret = ""
	if diff := cmp.Diff(app, gotApp, protocmp.Transform()); diff != "" {
		t.Errorf("have diff '%v'", diff)
	}
}
