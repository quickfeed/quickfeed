package database

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"
)

func TestTimestampSerializer_Value(t *testing.T) {
	type args struct {
		ctx        context.Context
		field      *schema.Field
		dst        reflect.Value
		fieldValue interface{}
	}
	tests := []struct {
		name    string
		tr      TimestampSerializer
		args    args
		want    interface{}
		wantErr error
	}{
		{
			name: "ok field",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: timestamppb.New(time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC)),
			},
			want:    time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
			wantErr: nil,
		},
		{
			name: "nil field",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: nil,
			},
			want:    nil,
			wantErr: nil,
		},
		{
			name: "field wrong type",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: "string",
			},
			want:    nil,
			wantErr: ErrUnsupportedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TimestampSerializer{}

			got, err := tr.Value(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.fieldValue)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("TimestampSerializer.Value() = %v, want %v", err, tt.wantErr)
			}
			if got == nil {
				return // ignore nil values
			}
			ts := &timestamppb.Timestamp{}
			if err = proto.Unmarshal(got.([]byte), ts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotTime := ts.AsTime()
			if !reflect.DeepEqual(gotTime, tt.want) {
				t.Errorf("TimestampSerializer.Value() = %v, want %v", gotTime, tt.want)
			}
		})
	}
}

func TestTimestampSerializer_Scan(t *testing.T) {
	type args struct {
		ctx     context.Context
		field   *schema.Field
		dst     reflect.Value
		dbValue interface{}
	}
	tests := []struct {
		name    string
		tr      TimestampSerializer
		args    args
		wantErr error
	}{
		{
			name: "db wrong type",
			args: args{
				field:   &schema.Field{},
				dbValue: "string",
			},
			wantErr: ErrUnsupportedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TimestampSerializer{}
			err := tr.Scan(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.dbValue)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("TimestampSerializer.Scan() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
