package auth

import (
	"os"

	"github.com/markbates/goth"
)

// TeacherSuffix is the suffix appended to the provider with the teacher scope.
const TeacherSuffix = "-teacher"

// Provider contains information about how to enable the same authentication
// provider with different scopes. The provider will be registered under Name
// with the student scope, and under Name + TeacherSuffix with the teacher
// scope.
type Provider struct {
	Name          string
	KeyEnv        string
	SecretEnv     string
	CallbackURL   string
	StudentScopes []string
	TeacherScopes []string
}

// EnableProvider enables the specified provider and returns true if the
// corresponding environment variables are set.
func EnableProvider(p *Provider, createProvider func(key, secret, callback string, scopes ...string) goth.Provider) bool {
	key := os.Getenv(p.KeyEnv)
	secret := os.Getenv(p.SecretEnv)
	if key == "" || secret == "" {
		return false
	}
	student := createProvider(key, secret, p.CallbackURL, p.StudentScopes...)
	student.SetName(p.Name)
	teacher := createProvider(key, secret, p.CallbackURL, p.TeacherScopes...)
	teacher.SetName(p.Name + TeacherSuffix)
	goth.UseProviders(student, teacher)
	return true
}
