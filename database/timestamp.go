package database

import (
	"context"
	"errors"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"
)

var ErrUnsupportedType = errors.New("unsupported type")

// TimestampSerializer is a GORM serializer that allows the serialization and deserialization of the
// google.protobuf.Timestamp protobuf message type.
type TimestampSerializer struct{}

// Value implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerValuerInterface to indicate
// how this struct will be saved into an SQL database field.
// Serializing timestamppb.Timestamp to time.Time allows saving it to database as "datetime" type.
func (TimestampSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}
	t, ok := fieldValue.(*timestamppb.Timestamp)
	if !ok {
		return nil, ErrUnsupportedType
	}
	if t == nil {
		// explicitly return nil to avoid saving empty timestamp as "0001-01-01T00:00:00Z"
		return nil, nil
	}
	return t.AsTime(), nil
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how
// this struct can be loaded from an SQL database field.
func (TimestampSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var ts *timestamppb.Timestamp
	if dbValue != nil {
		t, ok := dbValue.(time.Time)
		if !ok {
			return ErrUnsupportedType
		}
		ts = timestamppb.New(t)
		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(ts))
	}
	return
}
