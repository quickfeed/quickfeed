package qtest

import "github.com/quickfeed/quickfeed/qf"

const MockOrg = "qfTestOrg"

var MockCourses = []*qf.Course{
	{
		Name:             "Distributed Systems",
		CourseCreatorID:  1,
		Code:             "DAT520",
		Year:             2018,
		Tag:              "Spring",
		Provider:         "fake",
		OrganizationID:   1,
		OrganizationName: MockOrg,
	},
	{
		Name:             "Operating Systems",
		CourseCreatorID:  1,
		Code:             "DAT320",
		Year:             2017,
		Tag:              "Fall",
		Provider:         "fake",
		OrganizationID:   2,
		OrganizationName: "DAT320",
	},
	{
		Name:             "New Systems",
		CourseCreatorID:  1,
		Code:             "DATx20",
		Year:             2019,
		Tag:              "Fall",
		Provider:         "fake",
		OrganizationID:   3,
		OrganizationName: "DATx20-2019",
	},
	{
		Name:             "Hyped Systems",
		CourseCreatorID:  1,
		Code:             "DATx20",
		Year:             2020,
		Tag:              "Fall",
		Provider:         "fake",
		OrganizationID:   4,
		OrganizationName: "DATx20-2020",
	},
}
