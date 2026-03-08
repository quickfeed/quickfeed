package mockdata

import (
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
)

var mockOrg = &github.Organization{
	ID:    github.Int64(int64(1)),
	Login: github.String("qf-2026"),
}

func MockGroups() map[string]map[string][]github.User {
	groups := make(map[string]map[string][]github.User)
	for _, course := range defaultConfig.Courses {
		repo := github.Repository{
			Organization: mockOrg,
			Name:         github.String(course),
		}
		groupMembers := make([]github.User, len(Members))
		for i, member := range Members {
			groupMembers[i] = github.User{
				ID:   github.Int64(int64(i)),
				Name: github.String(member),
			}
		}
		groups[course] = map[string][]github.User{
			repo.GetName(): groupMembers,
		}
	}
	return groups
}

func MockRepos() []github.Repository {
	var rs []github.Repository
	for range defaultConfig.Courses {
		for _, r := range []string{"tests", "assignments", "info"} {
			rs = append(rs, github.Repository{
				Organization: mockOrg,
				Name:         github.String(r),
			})
		}
		for i, member := range Members {
			rs = append(rs, github.Repository{
				ID:           github.Int64(int64(i)),
				Organization: mockOrg,
				Name:         github.String(qf.StudentRepoName(member)),
			})
		}
		for i, group := range defaultConfig.Groups {
			rs = append(rs, github.Repository{
				ID:           github.Int64(int64(i)),
				Organization: mockOrg,
				Name:         github.String(group),
			})
		}
	}
	return rs
}
