package scm

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates that the user is not member of the organization.
	ErrNotMember = errors.New("not a member of the organization")
	// ErrNotOwner indicates that the user is not an owner of the organization.
	ErrNotOwner = errors.New("not an owner of the organization")
	// ErrAlreadyExists indicates that one or more QuickFeed repositories already exist in the organization.
	ErrAlreadyExists = errors.New("course repositories already exist")
)

// SCMError holds the operation, user error message and the original error.
type SCMError struct {
	op      Op
	userErr Msg
	err     error
}

var _ error = (*SCMError)(nil)

type (
	// Op describes an operation, such as "GetOrganization".
	Op string
	// Msg is the error message to be displayed to the user.
	Msg string
)

func M(format string, a ...interface{}) Msg {
	return Msg(fmt.Sprintf(format, a...))
}

func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to scm.E with no arguments")
	}
	e := &SCMError{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.op = arg
		case Msg:
			e.userErr = arg
		case *SCMError:
			e.err = arg
		case string:
			e.err = errors.New(arg)
		case error:
			e.err = arg
		}
	}
	return e
}

// Error returns the error message to be logged.
func (e *SCMError) Error() string {
	return fmt.Sprintf("scm.%s: %v", e.op, e.err)
}

func (e *SCMError) Unwrap() error {
	return e.err
}

// UserError returns the error message to be displayed to the user.
// It returns the first error in the chain of user errors.
func (e *SCMError) UserError() error {
	return errors.New(string(e.userErr))
}

// AllUserErrors returns all user errors in the error chain.
func (e *SCMError) AllUserErrors() error {
	err := errors.New(string(e.userErr))
	var se *SCMError
	for errors.As(e.err, &se) {
		err = fmt.Errorf("%s: %w", err, errors.New(string(se.userErr)))
		e = se
	}
	return err
}
