package qf

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

// IsOwner returns true if the current user is the same as the given user ID.
func (u *User) IsOwner(userID uint64) bool {
	return u.GetID() == userID
}

// UserIDs returns the user ID of this user.
func (u *User) UserIDs() []uint64 {
	return []uint64{u.GetID()}
}

// ValidateProfile checks that the user has complete profile information required for enrollment.
// It validates that Name (with at least first and last name), Email (valid format), and StudentID are set.
// Returns nil if valid, or an error describing the validation failure.
func (u *User) ValidateProfile() error {
	if u.GetName() == "" {
		return errors.New("name is required")
	}
	// Check that name has at least two components (first and last name)
	nameParts := strings.Fields(u.GetName())
	if len(nameParts) < 2 {
		return errors.New("name must contain at least first and last name")
	}
	if u.GetEmail() == "" {
		return errors.New("email is required")
	}
	// Validate that email is a proper email address
	if _, err := mail.ParseAddress(u.GetEmail()); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}
	if u.GetStudentID() == "" {
		return errors.New("student ID is required")
	}
	return nil
}
