package qtest

import "github.com/quickfeed/quickfeed/qf"

const MockOrg = "qf102-2022"

var MockCourses = []*qf.Course{
	{
		Name:                "Distributed Systems",
		CourseCreatorID:     1,
		Code:                "DAT520",
		Year:                2018,
		Tag:                 "Spring",
		ScmOrganizationID:   1,
		ScmOrganizationName: MockOrg,
	},
	{
		Name:                "Operating Systems",
		CourseCreatorID:     1,
		Code:                "DAT320",
		Year:                2017,
		Tag:                 "Fall",
		ScmOrganizationID:   2,
		ScmOrganizationName: "dat320",
	},
	{
		Name:                "New Systems",
		CourseCreatorID:     1,
		Code:                "DATx20",
		Year:                2019,
		Tag:                 "Fall",
		ScmOrganizationID:   3,
		ScmOrganizationName: "dat1020-2019",
	},
	{
		Name:                "Hyped Systems",
		CourseCreatorID:     1,
		Code:                "QF104",
		Year:                2022,
		Tag:                 "Fall",
		ScmOrganizationID:   4,
		ScmOrganizationName: "qf104-2022",
	},
}
