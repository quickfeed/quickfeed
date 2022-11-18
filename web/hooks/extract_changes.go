package hooks

import "strings"

func extractChanges(changes []string, modifiedAssignments map[string]bool) {
	for _, changedFile := range changes {
		// we assume the first path component holds the assignment name
		name, _, ok := strings.Cut(changedFile, "/")
		if !ok {
			// ignore root-level files
			continue
		}
		if name == "" {
			// ignore names that start with "/" or empty names
			continue
		}
		modifiedAssignments[name] = true
	}
}
