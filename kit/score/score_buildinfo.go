package score

import (
	"database/sql/driver"
	"fmt"

	"google.golang.org/protobuf/proto"
)

// TODO(meling) delete this file

func (b *BuildInfo) xScan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal BuildInfo value: %v", value)
	}
	return proto.Unmarshal(bytes, b)
}

func (b *BuildInfo) xValue() (driver.Value, error) {
	return proto.Marshal(b)
}

// // Scan scan value into Jsonb, implements sql.Scanner interface
// func (j *JSON) Scan(value interface{}) error {
//   bytes, ok := value.([]byte)
//   if !ok {
//     return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
//   }

//   result := json.RawMessage{}
//   err := json.Unmarshal(bytes, &result)
//   *j = JSON(result)
//   return err
// }

// // Value return json value, implement driver.Valuer interface
// func (j JSON) Value() (driver.Value, error) {
//   if len(j) == 0 {
//     return nil, nil
//   }
//   return json.RawMessage(j).MarshalJSON()
// }
