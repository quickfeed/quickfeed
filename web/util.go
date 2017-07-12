package web

import "strconv"

// ParseUintParam returns the uint of the provided string s.
// If the length of s is 0, then 0 is returned.
func ParseUintParam(s string) (uint64, error) {
	if len(s) == 0 {
		return 0, nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return n, err
	}
	return n, nil
}
