/* eslint-disable no-unused-vars */
import { create } from "@bufbuild/protobuf"
import { BuildInfoSchema, ScoreSchema } from "../../../proto/kit/score/score_pb"
import {
    CourseSubmissionsSchema,
    Organization,
    OrganizationSchema,
} from "../../../proto/qf/requests_pb"
import {
    Assignment,
    AssignmentSchema,
    CourseSchema,
    Enrollment,
    Enrollment_DisplayState,
    Enrollment_UserStatus,
    EnrollmentSchema,
    EnrollmentsSchema,
    GradeSchema,
    GradingBenchmark,
    GradingBenchmarkSchema,
    GradingCriterion_Grade,
    GradingCriterionSchema,
    Group_GroupStatus,
    GroupSchema,
    GroupsSchema,
    Repository_Type,
    Review,
    ReviewSchema,
    Submission_Status,
    SubmissionSchema,
    SubmissionsSchema,
    User,
    UserSchema,
} from "../../../proto/qf/types_pb"
import { SubmissionsForCourse } from "../../Helpers"
import { timestampFromDate } from "@bufbuild/protobuf/wkt"

export class MockData {
    public static mockedUsers(): User[] {
        const userList: User[] = []
        userList.push(
            create(UserSchema, {
                ID: BigInt(1),
                Name: "Test User",
                Email: "test@example.com",
                Login: "Test User",
                StudentID: "9999",
                IsAdmin: true,
            })
        )

        userList.push(
            create(UserSchema, {
                ID: BigInt(2),
                Name: "Admin Admin",
                Email: "admin@admin",
                Login: "Admin",
                StudentID: "1000",
                IsAdmin: true,
            })
        )

        userList.push(
            create(UserSchema, {
                ID: BigInt(3),
                Name: "Test Student",
                Email: "test@student.no",
                Login: "Student",
                AvatarURL: "https://avatars0.githubusercontent.com/u/1?v=4",
                IsAdmin: false,
            })
        )

        userList.push(
            create(UserSchema, {
                ID: BigInt(4),
                Name: "Bob Bobsen",
                Email: "bob@bobsen.no",
                Login: "Bob",
                StudentID: "1234",
                IsAdmin: true,
            })
        )

        userList.push(
            create(UserSchema, {
                ID: BigInt(5),
                Name: "Petter Pan",
                Email: "petter@pan.no",
                StudentID: "2345",
                IsAdmin: false,
            })
        )
        return userList
    }

    public static mockedAssignments(): Assignment[] {
        const ts = timestampFromDate(new Date(2017, 5, 25))
        const a0 = create(AssignmentSchema)
        const a1 = create(AssignmentSchema)
        const a2 = create(AssignmentSchema)
        const a3 = create(AssignmentSchema)
        const a4 = create(AssignmentSchema)
        const a5 = create(AssignmentSchema)
        const a6 = create(AssignmentSchema)
        const a7 = create(AssignmentSchema)
        const a8 = create(AssignmentSchema)
        const a9 = create(AssignmentSchema)
        const a10 = create(AssignmentSchema)

        a0.ID = BigInt(1)
        a0.CourseID = BigInt(1)
        a0.name = "Lab 1"
        a0.deadline = ts
        a0.scoreLimit = 80
        a0.order = 1

        a1.ID = BigInt(2)
        a1.CourseID = BigInt(1)
        a1.name = ("Lab 2")
        a1.deadline = ts
        a1.scoreLimit = 80
        a1.order = 2

        a2.ID = BigInt(3)
        a2.CourseID = BigInt(1)
        a2.name = "Lab 3"
        a2.reviewers = 1
        a2.deadline = ts
        a2.scoreLimit = 60
        a2.order = 3

        a3.ID = BigInt(4)
        a3.CourseID = BigInt(1)
        a3.name = "Lab 4"
        a3.deadline = ts
        a3.scoreLimit = 75
        a3.order = 4
        a3.isGroupLab = true

        a4.ID = BigInt(5)
        a4.CourseID = BigInt(2)
        a4.name = "Lab 1"
        a4.deadline = ts
        a4.scoreLimit = 90
        a4.order = 1

        a5.ID = BigInt(6)
        a5.CourseID = BigInt(2)
        a5.name = "Lab 2"
        a5.deadline = ts
        a5.scoreLimit = 85
        a5.order = 2

        a6.ID = BigInt(7)
        a6.CourseID = BigInt(2)
        a6.name = "Lab 3"
        a6.deadline = ts
        a6.scoreLimit = 80
        a6.order = 3

        a7.ID = BigInt(8)
        a7.CourseID = BigInt(3)
        a7.name = "Lab 1"
        a7.deadline = ts
        a7.scoreLimit = 90
        a7.order = 1

        a8.ID = BigInt(9)
        a8.CourseID = BigInt(3)
        a8.name = "Lab 2"
        a8.deadline = ts
        a8.scoreLimit = 85
        a8.order = 2

        a9.ID = BigInt(10)
        a9.CourseID = BigInt(4)
        a9.name = "Lab 1"
        a9.deadline = ts
        a9.scoreLimit = 90
        a9.order = 1

        a10.ID = BigInt(11)
        a10.CourseID = BigInt(5)
        a10.name = "Lab 1"
        a10.deadline = ts
        a10.scoreLimit = 90
        a10.order = 1

        return [a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10]
    }

    public static mockedCourseAssignments(): { [key: string]: Assignment[] } {
        const courseAssignments: { [key: string]: Assignment[] } = {}
        const assignments = MockData.mockedAssignments()
        for (const assignment of assignments) {
            if (courseAssignments[assignment.CourseID.toString()]) {
                courseAssignments[assignment.CourseID.toString()].push(assignment)
            } else {
                courseAssignments[assignment.CourseID.toString()] = [assignment]
            }
        }
        return courseAssignments
    }

    public static mockedCourses() {
        const course0 = create(CourseSchema)
        const course1 = create(CourseSchema)
        const course2 = create(CourseSchema)
        const course3 = create(CourseSchema)
        const course4 = create(CourseSchema)

        course0.ID = BigInt(1)
        course0.name = "Object Oriented Programming"
        course0.code = "DAT100"
        course0.tag = "Spring"
        course0.year = 2017
        course0.ScmOrganizationID = BigInt(23650610)
        course0.courseCreatorID = BigInt(1)

        course1.ID = BigInt(2)
        course1.name = "Algorithms and Data Structures"
        course1.code = "DAT200"
        course1.tag = "Spring"
        course1.year = 2017
        course1.ScmOrganizationID = BigInt(23650611)

        course2.ID = BigInt(3)
        course2.name = "Databases"
        course2.code = "DAT220"
        course2.tag = "Spring"
        course2.year = 2017
        course2.ScmOrganizationID = BigInt(23650612)

        course3.ID = BigInt(4)
        course3.name = "Communication Technology"
        course3.code = "DAT230"
        course3.tag = "Spring"
        course3.year = 2017
        course3.ScmOrganizationID = BigInt(23650613)

        course4.ID = BigInt(5)
        course4.name = "Operating Systems"
        course4.code = "DAT320"
        course4.tag = "Spring"
        course4.year = 2017
        course4.ScmOrganizationID = BigInt(23650614)

        return [course0, course1, course2, course3, course4]
    }

    public static mockedEnrollments() {
        const enrollments = create(EnrollmentsSchema)
        const localEnrols: Enrollment[] = []
        localEnrols.push(
            create(EnrollmentSchema, {
                ID: BigInt(1),
                courseID: BigInt(1),
                userID: BigInt(1),
                status: Enrollment_UserStatus.TEACHER,
                state: Enrollment_DisplayState.VISIBLE,
                groupID: BigInt(1),
                user: MockData.mockedUsers().find((u) => u.ID === BigInt(1)),
            })
        )

        localEnrols.push(
            create(EnrollmentSchema, {
                ID: BigInt(2),
                courseID: BigInt(2),
                userID: BigInt(1),
                status: Enrollment_UserStatus.TEACHER,
                state: Enrollment_DisplayState.VISIBLE,
                user: MockData.mockedUsers().find((u) => u.ID === BigInt(1)),
            })
        )

        localEnrols.push(
            create(EnrollmentSchema, {
                ID: BigInt(3),
                courseID: BigInt(1),
                userID: BigInt(2),
                status: Enrollment_UserStatus.STUDENT,
                groupID: BigInt(1),
                user: MockData.mockedUsers().find((u) => u.ID === BigInt(2)),
            })
        )

        localEnrols.push(
            create(EnrollmentSchema, {
                ID: BigInt(4),
                courseID: BigInt(2),
                userID: BigInt(2),
                status: Enrollment_UserStatus.PENDING,
                user: MockData.mockedUsers().find((u) => u.ID === BigInt(2)),
            })
        )

        localEnrols.push(
            create(EnrollmentSchema, {
                ID: BigInt(5),
                courseID: BigInt(1),
                userID: BigInt(3),
                status: Enrollment_UserStatus.STUDENT,
                groupID: BigInt(2),
                user: MockData.mockedUsers().find((u) => u.ID === BigInt(3)),
            })
        )

        localEnrols.push(
            create(EnrollmentSchema, {
                ID: BigInt(6),
                courseID: BigInt(1),
                userID: BigInt(4),
                status: Enrollment_UserStatus.STUDENT,
                groupID: BigInt(2),
                user: MockData.mockedUsers().find((u) => u.ID === BigInt(4)),
            })
        )
        enrollments.enrollments = (localEnrols)
        return enrollments
    }

    public static mockedOrganizations(): Organization[] {
        const localOrgs: Organization[] = []
        const localOrg = create(OrganizationSchema)
        localOrg.ScmOrganizationID = BigInt(23650610)
        localOrg.ScmOrganizationName = "test"
        localOrgs.push(localOrg)
        return localOrgs
    }

    public static mockedGroups() {
        const groups = create(GroupsSchema)

        const group1 = create(GroupSchema, {
            ID: BigInt(1),
            name: "Group 1",
            status: Group_GroupStatus.APPROVED,
            courseID: BigInt(1),
        })

        const group2 = create(GroupSchema, {
            ID: BigInt(2),
            name: "Group 2",
            status: Group_GroupStatus.PENDING,
            courseID: BigInt(1),
        })

        groups.groups = [group1, group2]
        return groups
    }

    public static mockedSubmissions() {
        const submissions = create(SubmissionsSchema)
        submissions.submissions = [
            create(SubmissionSchema, {
                ID: BigInt(1),
                AssignmentID: BigInt(1),
                userID: BigInt(1),
                Grades: [
                    create(GradeSchema, {
                        Status: Submission_Status.APPROVED,
                        SubmissionID: BigInt(1),
                        UserID: BigInt(1),
                    })
                ],
                BuildInfo: create(BuildInfoSchema, {
                    ID: BigInt(1),
                    SubmissionID: BigInt(1),
                    ExecTime: BigInt(1),
                    BuildDate: timestampFromDate(new Date(2017, 6, 4)),
                    SubmissionDate: timestampFromDate(new Date(2017, 6, 4)),
                    BuildLog: "Build log for submission 1",
                }),
                score: 100,
                commitHash: "abc",
                Scores: [
                    create(ScoreSchema, {
                        ID: BigInt(1),
                        SubmissionID: BigInt(1),
                        MaxScore: 10,
                        Score: 10,
                        TestName: "Test 1",
                        Weight: 2
                    }),
                    create(ScoreSchema, {
                        ID: BigInt(2),
                        SubmissionID: BigInt(1),
                        MaxScore: 10,
                        Score: 10,
                        TestName: "Test 2",
                        Weight: 2
                    }),
                ],
            }),

            create(SubmissionSchema, {
                ID: BigInt(2),
                AssignmentID: BigInt(2),
                userID: BigInt(2),
                score: 75,
                commitHash: "bcd",
            }),

            create(SubmissionSchema, {
                ID: BigInt(3),
                AssignmentID: BigInt(3),
                userID: BigInt(1),
                score: 80,
                released: true,
                reviews: [
                    create(ReviewSchema, {
                        ID: BigInt(1),
                        SubmissionID: BigInt(3),
                        score: 80,
                        feedback: "Well done!",
                        ReviewerID: BigInt(1),
                        gradingBenchmarks: [
                            create(GradingBenchmarkSchema, {
                                ID: BigInt(1),
                                AssignmentID: BigInt(2),
                                heading: "HTML",
                                ReviewID: BigInt(1),
                                criteria: [
                                    create(GradingCriterionSchema, {
                                        ID: BigInt(1),
                                        BenchmarkID: BigInt(1),
                                        description: "Add div",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    }),
                                    create(GradingCriterionSchema, {
                                        ID: BigInt(2),
                                        BenchmarkID: BigInt(1),
                                        description: "Div has text",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    })
                                ]
                            }),
                            create(GradingBenchmarkSchema, {
                                ID: BigInt(2),
                                AssignmentID: BigInt(2),
                                heading: "CSS",
                                ReviewID: BigInt(1),
                                criteria: [
                                    create(GradingCriterionSchema, {
                                        ID: BigInt(3),
                                        BenchmarkID: BigInt(2),
                                        description: "Div centered",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    }),
                                    create(GradingCriterionSchema, {
                                        ID: BigInt(4),
                                        BenchmarkID: BigInt(2),
                                        description: "Div colored",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    })
                                ]
                            })
                        ]
                    }),
                ]
            }),
            create(SubmissionSchema, {
                ID: BigInt(4),
                AssignmentID: BigInt(4),
                groupID: BigInt(1),
                score: 90,
                commitHash: "def",
            }),
            create(SubmissionSchema, {
                ID: BigInt(5),
                AssignmentID: BigInt(5),
                userID: BigInt(1),
                score: 100,
                commitHash: "efg",
            }),

            create(SubmissionSchema, {
                ID: BigInt(6),
                AssignmentID: BigInt(1),
                userID: BigInt(3),
                score: 50,
                commitHash: "test",
                Grades: [
                    create(GradeSchema, {
                        Status: Submission_Status.NONE,
                        SubmissionID: BigInt(6),
                        UserID: BigInt(3),
                    })
                ],
                BuildInfo: create(BuildInfoSchema, {
                    ID: BigInt(3),
                    BuildDate: timestampFromDate(new Date(2022, 6, 4)),
                    BuildLog: "Build log for test student",
                    ExecTime: BigInt(1),
                }),
                Scores: [
                    create(ScoreSchema, {
                        ID: BigInt(3),
                        MaxScore: 10,
                        Score: 5,
                        SubmissionID: BigInt(6),
                        TestName: "Test 1",
                        TestDetails: "Test details for test 1",
                        Weight: 5,
                    }),
                    create(ScoreSchema, {
                        ID: BigInt(4),
                        MaxScore: 10,
                        Score: 7,
                        SubmissionID: BigInt(6),
                        TestName: "Test 2",
                        TestDetails: "Test details for test 2",
                        Weight: 2,
                    }),
                ]
            }),
            create(SubmissionSchema, {
                ID: BigInt(7),
                AssignmentID: BigInt(4),
                userID: BigInt(2),
                score: 75,
                commitHash: "bcd",
            }),

        ]
        return submissions
    }

    public static mockedCourseSubmissions(courseID: bigint): SubmissionsForCourse {
        const userSubmissions = create(CourseSubmissionsSchema)
        const groupSubmissions = create(CourseSubmissionsSchema)

        const assignments = MockData.mockedAssignments().filter((a) => a.CourseID === courseID)
        const submissions = MockData.mockedSubmissions().submissions.filter((s) => assignments.map((a) => a.ID).includes(s.AssignmentID))
        const enrollments = MockData.mockedEnrollments().enrollments.filter((e) => e.courseID === courseID)
        const groups = MockData.mockedGroups().groups.filter((g) => g.courseID === courseID)
        const sfc = new SubmissionsForCourse()
        for (const enrollment of enrollments) {
            const subs = submissions.filter((s) => s.userID === enrollment.userID)
            userSubmissions.submissions[enrollment.ID.toString()] = create(SubmissionsSchema, { submissions: subs })
        }

        for (const group of groups) {
            const groupSubs = submissions.filter((s) => s.groupID === group.ID)
            groupSubmissions.submissions[group.ID.toString()] = create(SubmissionsSchema, { submissions: groupSubs })
        }

        sfc.setSubmissions("USER", userSubmissions)
        sfc.setSubmissions("GROUP", groupSubmissions)
        return sfc
    }

    public static mockedBenchmarks(): GradingBenchmark[] {
        const templateBenchmarks = []

        templateBenchmarks.push(
            create(GradingBenchmarkSchema, {
                ID: BigInt(1),
                AssignmentID: BigInt(1),
                heading: "HTML",
                criteria: [
                    create(GradingCriterionSchema, {
                        ID: BigInt(1),
                        BenchmarkID: BigInt(1),
                        description: "Add div",
                        points: BigInt(10),
                    }),
                    create(GradingCriterionSchema, {
                        ID: BigInt(2),
                        BenchmarkID: BigInt(1),
                        description: "Div has text",
                        points: BigInt(10),
                    }),
                ]
            }),
            create(GradingBenchmarkSchema, {
                ID: BigInt(2),
                AssignmentID: BigInt(2),
                heading: "CSS",
                criteria: [
                    create(GradingCriterionSchema, {
                        ID: BigInt(3),
                        BenchmarkID: BigInt(2),
                        description: "Div centered",
                        points: BigInt(10),
                    }),
                    create(GradingCriterionSchema, {
                        ID: BigInt(4),
                        BenchmarkID: BigInt(2),
                        description: "Div colored",
                        points: BigInt(10),
                    }),
                ]
            })
        )
        return templateBenchmarks
    }

    public static mockedRepositories() {
        const repositories: { [courseid: string]: { [repo: number]: string } } = {
            "1": {
                [Repository_Type.INFO]: "info",
                [Repository_Type.ASSIGNMENTS]: "assignments",
                [Repository_Type.USER]: "user",
                [Repository_Type.GROUP]: "group",
                [Repository_Type.TESTS]: "tests",
            }
        }
        return repositories
    }
    public static computeScore(r: Review) {
        let score = 0
        let totalApproved = 0
        let total = 0
        for (let i = 0; i < r.gradingBenchmarks.length; i++) {
            const gb = r.gradingBenchmarks[i]
            for (let j = 0; j < gb.criteria.length; j++) {
                const criterion = gb.criteria[j]
                total++
                if (criterion.grade === GradingCriterion_Grade.PASSED) {
                    score += Number(criterion.points)
                    totalApproved++
                }
            }
        }
        if (score === 0) {
            score = 100 / total * totalApproved
        }
        return score
    }
}
