package multierr

import (
	"fmt"
	"strings"
)

type multiErr []error

func (errs multiErr) Error() string {
	var b strings.Builder
	if len(errs) == 1 {
		return errs[0].Error()
	}
	for _, e := range errs {
		fmt.Fprintf(&b, "%s\n", e)
	}
	return b.String()
}

func Join(errs ...error) error {
	err := make(multiErr, 0, len(errs))
	for _, e := range errs {
		if e == nil {
			continue
		}
		err = append(err, e)
	}
	if len(err) == 0 {
		return nil
	}
	return err
}
