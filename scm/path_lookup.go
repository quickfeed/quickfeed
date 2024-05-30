package scm

import (
	"fmt"
	"strconv"
	"strings"
)

func lookup(key, pattern, url string) string {
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

func lookupInt(key, pattern, url string) (int, error) {
	// Use the lookup function to get the string value
	value := lookup(key, pattern, url)
	if value == "" {
		return 0, fmt.Errorf("no value found for key %s", key)
	}
	// Convert the string value to an integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("could not convert value to int: %v", err)
	}
	return intValue, nil
}
