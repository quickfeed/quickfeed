package web

import (
	"strings"

	pb "github.com/autograde/aguis/ag"

	"github.com/markbates/goth"
)

// TeacherSuffix is used to set user as a teacher to be able to create new course
const TeacherSuffix = "-teacher"

// GetProviders returns a list of all providers enabled by goth
func GetProviders() *pb.Providers {
	var providers []string
	for _, provider := range goth.GetProviders() {
		if !strings.HasSuffix(provider.Name(), TeacherSuffix) {
			providers = append(providers, provider.Name())
		}
	}
	return &pb.Providers{Providers: providers}
}
