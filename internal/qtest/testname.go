package qtest

import (
	"fmt"
	"strings"
)

// Name returns the test name based on the provided fields and values.
func Name(name string, fields []string, values ...any) string {
	if len(fields) != len(values) {
		panic("fields and values must have the same length")
	}
	b := strings.Builder{}
	b.WriteString(name)
	for i, f := range fields {
		v := values[i]
		switch x := v.(type) {
		case []string:
			if x == nil {
				fmt.Fprintf(&b, "/%s=<nil>", f)
			} else {
				fmt.Fprintf(&b, "/%s=%v", f, x)
			}
		case string:
			if x != "" {
				fmt.Fprintf(&b, "/%s=%s", f, v)
			}
		case uint64:
			if x != 0 {
				fmt.Fprintf(&b, "/%s=%d", f, v)
			}
		case int:
			if x != 0 {
				fmt.Fprintf(&b, "/%s=%d", f, v)
			}
		case bool:
			if x {
				fmt.Fprintf(&b, "/%s", f)
			}
		default:
			fmt.Fprintf(&b, "/%s=%v", f, v)
		}
	}
	return b.String()
}
