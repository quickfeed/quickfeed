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
func (TimestampSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}
	t, ok := fieldValue.(*timestamppb.Timestamp)
	if !ok {
		return nil, ErrUnsupportedType
	}
	return t.AsTime(), nil
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how
// this struct can be loaded from an SQL database field.
func (TimestampSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var ts *timestamppb.Timestamp
	if dbValue != nil {
		switch value := dbValue.(type) {
		case time.Time:
			ts = timestamppb.New(value)
		default:
			return ErrUnsupportedType
		}
		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(ts))
	}
	return
}
