package scm

import (
	"strconv"
	"strings"
)

func pathValue(key, pattern, url string) string {
	urlParts := strings.Split(url, "/")

	// find the key in pattern
	for i, part := range strings.Split(pattern, "/") {
		if strings.Contains(part, "{"+key+"}") {
			if len(urlParts) > i {
				return urlParts[i]
			}
			break // found key in pattern, but no value in URL; return empty string.
		}
	}
	return ""
}

// mustParseInt returns the integer value of the key in the URL.
// If the key is not found, or the value cannot be converted to an integer, the function panics.
func mustParseInt(key, pattern, url string) int {
	// Use the lookup function to get the string value
	value := pathValue(key, pattern, url)
	if value == "" {
		panic("no value found for key")
	}
	// Convert the string value to an integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		// could not convert value to int
		panic(err)
	}
	return intValue
}
