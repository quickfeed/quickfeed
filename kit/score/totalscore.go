package score

// Validate returns an error if one of the recorded score objects are invalid.
// Otherwise, nil is returned.
func Validate() error {
	for _, sc := range scores {
		if err := sc.IsValid(""); err != nil {
			return err
		}
	}
	return nil
}
