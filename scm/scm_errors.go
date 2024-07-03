package scm

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates that the user is not member of the organization.
	ErrNotMember = errors.New("not a member of organization")
	// ErrNotOwner indicates that the user is not an owner of the organization.
	ErrNotOwner = errors.New("not an owner of organization")
	// ErrAlreadyExists indicates that a repository already exist in the organization.
	ErrAlreadyExists = errors.New("already exist")
)

// SCMError holds the operation, user error message and the original error.
type SCMError struct {
	op  Op
	err error
}

type unwrap interface{ Unwrap() error }

var (
	_ error  = (*SCMError)(nil)
	_ unwrap = (*SCMError)(nil)
	_ error  = (*UserError)(nil)
	_ unwrap = (*UserError)(nil)
)

// Op describes an operation, such as "GetOrganization".
type Op string

// UserError is an error that is meant to be displayed to the user.
type UserError struct {
	e error
}

func (e *UserError) Error() string {
	return e.e.Error()
}

func (e *UserError) Unwrap() error {
	return e.e
}

// M creates a new user error with the given format string.
func M(format string, a ...interface{}) error {
	return &UserError{fmt.Errorf(format, a...)}
}

// E creates a new SCM error with the given operation, error, and user error.
// The error message is constructed as "scm.<op>: <err>".
// The user error can be constructed with the M function.
// If more than one errors are passed, these are chained.
// If no arguments are passed, E panics.
func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to scm.E with no arguments")
	}
	e := &SCMError{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.op = arg
		case *UserError:
			e.add(arg)
		case *SCMError:
			e.add(arg)
		case string:
			e.add(errors.New(arg))
		case error:
			e.add(arg)
		}
	}
	return e
}

func (e *SCMError) add(err error) {
	if e.err == nil {
		e.err = err
	} else {
		e.err = fmt.Errorf("%w: %w", e.err, err)
	}
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
	var ue *UserError
	if errors.As(e.err, &ue) {
		return ue
	}
	return nil
}
