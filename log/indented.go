package log

import (
	"encoding/json"
	"fmt"
)

// IndentJson returns a JSON formatted string
// with structured indents and line breaks.
func IndentJson(event interface{}) string {
	prettyJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON error: %v", err)
	}
	return string(prettyJSON)
}
