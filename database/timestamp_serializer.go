package database

import (
	"context"
	"errors"
	"reflect"

	"google.golang.org/protobuf/proto"
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
	return proto.Marshal(t)
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how
// this struct can be loaded from an SQL database field.
func (TimestampSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	if dbValue != nil {
		b, ok := dbValue.([]byte)
		if !ok {
			return ErrUnsupportedType
		}
		t := &timestamppb.Timestamp{}
		if err = proto.Unmarshal(b, t); err != nil {
			return err
		}
		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(t))
	}
	return
}
