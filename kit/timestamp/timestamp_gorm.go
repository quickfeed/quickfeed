package timestamp

import (
	"database/sql/driver"

	"google.golang.org/protobuf/proto"
)

// Scan unmarshals value of type []byte from GORM into ts.
func (ts *Timestamp) Scan(value interface{}) error {
	return proto.Unmarshal(value.([]byte), ts)
}

// Value marshals ts into []byte value to be stored using GORM.
func (ts *Timestamp) Value() (driver.Value, error) {
	return proto.Marshal(ts)
}
