package score

import (
	"database/sql/driver"

	"google.golang.org/protobuf/proto"
)

// Scan scan value into Results, implements sql.Scanner interface
func (b *Results) Scan(value interface{}) error {
	return proto.Unmarshal(value.([]byte), b)
}

// Value return Results value, implement driver.Valuer interface
func (b *Results) Value() (driver.Value, error) {
	return proto.Marshal(b)
}

// Scan scan value into BuildInfo, implements sql.Scanner interface
// func (b *BuildInfo) Scan(value interface{}) error {
// 	return proto.Unmarshal(value.([]byte), b)
// }

// Value return BuildInfo value, implement driver.Valuer interface
// func (b *BuildInfo) Value() (driver.Value, error) {
// 	return proto.Marshal(b)
// }
