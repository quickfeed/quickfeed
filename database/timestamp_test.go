package database_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"
)

func TestTimestampSerializer_Value(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	tests := []struct {
		name       string
		fieldValue interface{}
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "correct timestamp, current time",
			fieldValue: timestamppb.New(now),
			want:       now,
			wantErr:    false,
		},
		{
			name:       "correct timestamp, preset time",
			fieldValue: timestamppb.New(time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC)),
			want:       time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
			wantErr:    false,
		},
		{
			name:       "incorrect type: time.Time",
			fieldValue: time.Now(),
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "incorrect type: string",
			fieldValue: "2022-11-11T23:59:00",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "empty interface",
			fieldValue: nil,
			want:       nil,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := database.TimestampSerializer{}
			got, err := ts.Value(ctx, &schema.Field{}, reflect.Value{}, tt.fieldValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got = %v, ", tt.name, tt.wantErr, err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch timestamp (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTimestampSerializer_Nil_Value(t *testing.T) {
	ctx := context.Background()
	assignment := &qf.Assignment{
		Order:    1,
		CourseID: 1,
		// Deadline is nil
	}
	ts := database.TimestampSerializer{}
	got, _ := ts.Value(ctx, &schema.Field{}, reflect.Value{}, assignment.Deadline)
	if got != nil {
		t.Errorf("expected nil, got = %v, ", got)
	}
}

func TestTimestampSerializer_Scan(t *testing.T) {
	ctx := context.Background()
	ts := database.TimestampSerializer{}
	tests := []struct {
		name    string
		field   *schema.Field
		dst     reflect.Value
		dbValue interface{}
		wantErr bool
	}{
		{
			name:    "incorrect db type",
			field:   &schema.Field{},
			dst:     reflect.Value{},
			dbValue: "2022-01-24 14:03:00 +0000 UTC",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.Scan(ctx, tt.field, tt.dst, tt.dbValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected error: %v, got = %v, ", tt.name, tt.wantErr, err)
			}
		})
	}
}
