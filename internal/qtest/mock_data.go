package qtest

import (
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/qf"
)

const MockOrg = "qf102-2022"

var (
	DAT520 = "Distributed Systems"
	DAT320 = "Operating Systems"
	DATx20 = "New Systems"
	QF104  = "Hyped Systems"
)

var MockCourses = []*qf.Course{
	{
		Name:                DAT520,
		CourseCreatorID:     1,
		Code:                "DAT520",
		Year:                2018,
		Tag:                 "Spring",
		ScmOrganizationID:   1,
		ScmOrganizationName: MockOrg,
	},
	{
		Name:                DAT320,
		CourseCreatorID:     1,
		Code:                "DAT320",
		Year:                2017,
		Tag:                 "Fall",
		ScmOrganizationID:   2,
		ScmOrganizationName: "dat320",
	},
	{
		Name:                DATx20,
		CourseCreatorID:     1,
		Code:                "DATx20",
		Year:                2019,
		Tag:                 "Fall",
		ScmOrganizationID:   3,
		ScmOrganizationName: "dat1020-2019",
	},
	{
		Name:                QF104,
		CourseCreatorID:     1,
		Code:                "QF104",
		Year:                2022,
		Tag:                 "Fall",
		ScmOrganizationID:   4,
		ScmOrganizationName: "qf104-2022",
	},
}

// Members is a list of members for each course.
var Members = []string{
	"Abigail Dyer",
	"Max Clarkson",
	"Irene Powell",
	"Peter Jones",
	"Jack Wallace",
	"Dominic Grant",
	"Amelia Henderson",
	"Sonia Rutherford",
	"Jack Marshall",
	"Jennifer Howard",
	"Carol Langdon",
	"Zoe Chapman",
	"Tracey Vaughan",
	"Peter Alsop",
	"Felicity Vaughan",
	"Oliver Welch",
	"Tim Harris",
	"Vanessa Carr",
	"Felicity Dowd",
	"Frank Rees",
	"Joan Ogden",
	"Alexandra Manning",
	"Kevin Underwood",
	"Gavin Howard",
	"Edward Dowd",
	"Steven Peters",
	"Liam Metcalfe",
	"Trevor Howard",
	"Dylan Miller",
	"Stephanie Ball",
	"Angela Poole",
	"Samantha Slater",
	"Thomas Smith",
	"Justin Clarkson",
	"Piers Cameron",
	"Carolyn Hodges",
	"William Wright",
	"Christian Lawrence",
	"Amy Kelly",
	"Bella Dickens",
	"Lauren Manning",
	"Tracey Bailey",
	"Gavin Stewart",
	"Brandon Stewart",
	"Diane Lyman",
	"Joan Roberts",
	"Anna Hill",
	"Samantha Ross",
	"Michael McLean",
	"David Peters",
	"Simon Reid",
	"Andrew Greene",
	"Diana MacLeod",
	"Edward Murray",
	"Molly MacDonald",
	"Donna Walsh",
	"Alan Taylor",
	"Ruth Wallace",
	"Madeleine Fisher",
	"Una Springer",
	"Cameron Slater",
	"Peter Russell",
	"Molly Black",
	"Grace McLean",
	"Bernadette Lawrence",
	"Warren Johnston",
	"Wendy Parsons",
	"Steven Watson",
	"Penelope Berry",
	"Blake Powell",
	"Theresa Wilson",
	"Zoe Mackenzie",
	"Samantha Tucker",
	"Brian Dickens",
	"Thomas Terry",
	"Stephen Rutherford",
	"Molly Brown",
	"Anna Wright",
	"Jacob Henderson",
	"Irene Davies",
	"Leah Simpson",
	"Una Fisher",
	"Molly Ogden",
	"Victor Sutherland",
	"Dominic Peters",
	"Lucas Ross",
	"Gavin Wilson",
	"Stephen Bailey",
	"Diana Lee",
	"Phil Thomson",
	"Alan Dickens",
	"Vanessa James",
	"Pippa Watson",
	"Anne Lawrence",
	"Anne Underwood",
	"Anna White",
	"Heather Clark",
	"Karen Hudson",
	"Una Powell",
	"Owen Vance",
}

var Groups = []string{
	"Alpha Team",
	"Beta Squad",
	"Gamma Group",
	"Delta Force",
	"Epsilon Circle",
	"Zeta Unit",
	"Eta Crew",
	"Theta Alliance",
	"Iota Network",
	"Kappa League",
	"Lambda Syndicate",
}

func MockGroups() map[string]map[string][]github.User {
	groups := make(map[string]map[string][]github.User)
	for _, course := range MockCourses {
		repo := github.Repository{
			ID:           github.Int64(int64(course.ScmOrganizationID)),
			Organization: toOrg(course),
			Name:         github.String("mock-repo"),
		}
		groupMembers := make([]github.User, len(Members))
		for i, member := range Members {
			groupMembers[i] = github.User{
				ID:   github.Int64(int64(i)),
				Name: github.String(member),
			}
		}
		groups[course.ScmOrganizationName] = map[string][]github.User{
			repo.GetName(): groupMembers,
		}
	}
	return groups
}

func toOrg(course *qf.Course) *github.Organization {
	return &github.Organization{
		ID:    github.Int64(int64(course.GetScmOrganizationID())),
		Login: github.String(course.GetScmOrganizationName()),
	}
}

func MockRepos() []github.Repository {
	var repos []github.Repository
	rs := []string{"tests", "assignments", "info"}
	for _, course := range MockCourses {
		org := toOrg(course)
		for _, r := range rs {
			repos = append(repos, github.Repository{
				Organization: org,
				Name:         github.String(r),
			})
		}
		for i, member := range Members {
			repos = append(repos, github.Repository{
				ID:           github.Int64(int64(i)),
				Organization: org,
				Name:         github.String(qf.StudentRepoName(member)),
			})
		}
		for i, group := range Groups {
			repos = append(repos, github.Repository{
				ID:           github.Int64(int64(i)),
				Organization: org,
				Name:         github.String(group),
			})
		}
	}
	return repos
}
