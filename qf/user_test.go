package qf

import (
	"strings"
	"testing"
)

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		wantErr     bool
		errContains string
	}{
		{
			name: "valid user with complete information",
			user: &User{
				Login:     "testuser",
				Name:      "Test User",
				Email:     "test@example.com",
				StudentID: "123456",
			},
			wantErr: false,
		},
		{
			name: "valid user with multiple name parts",
			user: &User{
				Login:     "johndoe",
				Name:      "John Middle Doe",
				Email:     "john@example.com",
				StudentID: "654321",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			user: &User{
				Login:     "testuser",
				Name:      "",
				Email:     "test@example.com",
				StudentID: "123456",
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "name with only one part",
			user: &User{
				Login:     "testuser",
				Name:      "SingleName",
				Email:     "test@example.com",
				StudentID: "123456",
			},
			wantErr:     true,
			errContains: "name must contain at least first and last name",
		},
		{
			name: "missing email",
			user: &User{
				Login:     "testuser",
				Name:      "Test User",
				Email:     "",
				StudentID: "123456",
			},
			wantErr:     true,
			errContains: "email is required",
		},
		{
			name: "invalid email format",
			user: &User{
				Login:     "testuser",
				Name:      "Test User",
				Email:     "not-an-email",
				StudentID: "123456",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "invalid email format with @",
			user: &User{
				Login:     "testuser",
				Name:      "Test User",
				Email:     "@example.com",
				StudentID: "123456",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "missing student ID",
			user: &User{
				Login:     "testuser",
				Name:      "Test User",
				Email:     "test@example.com",
				StudentID: "",
			},
			wantErr:     true,
			errContains: "student ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.ValidateProfile()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateProfile() expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateProfile() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateProfile() unexpected error = %v", err)
				}
			}
		})
	}
}
