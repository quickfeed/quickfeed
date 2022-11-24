/* eslint-disable no-unused-vars */
import {
    Assignments,
    Course,
    Courses,
    Enrollment,
    Enrollments,
    GradingBenchmark,
    GradingCriterion,
    Group,
    Groups,
    Review,
    Submission,
    Submissions,
    User,
    Users,
    Assignment,
    EnrollmentLink,
    SubmissionLink,
    Enrollment_UserStatus,
    Group_GroupStatus,
    Submission_Status,
    GradingCriterion_Grade,
    Repository_Type,
    Enrollment_DisplayState,
} from "../proto/qf/types_pb"
import {
    CourseSubmissions,
    Organization,
    Repositories,
    Status,
    Void,
    SubmissionRequest_SubmissionType,
} from "../proto/qf/requests_pb"
import { delay } from "./Helpers"
import { BuildInfo, Score } from "../proto/kit/score/score_pb"
import { Code } from "@bufbuild/connect-web"

export interface IGrpcResponse<T> {
    status: Status
    data?: T
}

/** The Generate enum contains the types we generate IDs for.
    This is used to keep track of the next ID to use for each type.
    @example this.idMap.get(Generate.Course) // returns the previously generated ID for Course
    @example this.generateID(Generate.Course) // returns the next ID to use for Course
 */
enum Generate {
    User = "user",
    Course = "course",
    Assignment = "assignment",
    Group = "group",
    Enrollment = "enrollment",
    Submission = "submission",
    Review = "review",
    Score = "score",
    BuildInfo = "buildInfo",
    Organization = "organization",
    Provider = "provider",
    Repository = "repository",
    GradingCriterion = "gradingCriterion",
    GradingBenchmark = "gradingBenchmark",
    TemplateCriterion = "templateCriterion",
    TemplateBenchmark = "templateBenchmark",
}

export class MockGrpcManager {

    constructor(id?: number) {
        this.initUsers()
        this.initAssignments()
        this.initCourses()
        this.initOrganizations()
        this.addLocalCourseGroups()
        this.addLocalCourseStudent()
        this.addLocalLabInfo()
        this.initBenchmarks()
        // By default, set the current user to the first user
        // Optionally, set the current user to a specific user
        // By passing -1, the current user will be set to null
        if (id) {
            this.setCurrentUser(id)
        } else {
            this.setCurrentUser(1)
        }
    }


    private groups: Groups
    private users: Users
    private enrollments: Enrollments
    private currentUser: User | null
    private assignments: Assignments
    private courses: Courses
    private organizations: Organization[]
    private submissions: Submissions
    private templateBenchmarks: GradingBenchmark[]
    // idMap is a map of auto incrementing IDs
    public idMap: Map<string, number> = new Map<string, number>()
    /** generate holds the available types we generate IDs for */
    public generate: typeof Generate = Generate

    public getMockedUsers() {
        return this.users
    }

    public setCurrentUser(id: number) {
        const user = this.users.users.find(u => Number(u.ID) === id)
        if (user) {
            this.currentUser = user
        } else {
            this.currentUser = null
        }
    }

    public getUser(): Promise<IGrpcResponse<User>> {
        return this.grpcSend<User>(this.currentUser)
    }

    public getUsers(): Promise<IGrpcResponse<Users>> {
        if (this.currentUser?.IsAdmin) {
            return this.grpcSend<Users>(this.users)
        }
        return this.grpcSend<Users>(null)
    }

    public updateUser(user: User): Promise<IGrpcResponse<Void>> {
        if (!this.currentUser?.IsAdmin) {
            return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unauthenticated) }))
        }
        const usr = this.users.users?.findIndex(u => u.ID === user.ID)
        if (usr > -1) {
            Object.assign(this.users.users[usr], user)
        }
        return this.grpcSend<Void>(new Void())
    }

    // /* COURSES */ //

    public createCourse(course: Course): Promise<IGrpcResponse<Course>> {
        let data: Course | null = null
        const found = this.courses.courses.find(c => c.ID === course.ID)
        const IsAdmin = this.currentUser?.IsAdmin
        const user = this.currentUser
        if (!found && user && IsAdmin) {
            course.ID = this.generateID(Generate.Course)
            course.courseCreatorID = user.ID

            this.courses.courses.push(course)

            // Create new enrollment
            const enrollment = new Enrollment()
            enrollment.courseID = course.ID
            enrollment.userID = user.ID
            enrollment.status = Enrollment_UserStatus.TEACHER
            enrollment.ID = (this.generateID(Generate.Enrollment))
            enrollment.course = course
            enrollment.user = user
            enrollment.slipDaysRemaining = course.slipDays
            this.enrollments.enrollments.push(enrollment)

            data = course
        }
        return this.grpcSend<Course>(data)
    }

    public updateCourse(course: Course): Promise<IGrpcResponse<Void>> {
        const courseID = course.ID
        const courseIndex = this.courses.courses.findIndex(c => c.ID === courseID)
        if (courseIndex > -1) {
            const courses = this.courses.courses
            Object.assign(courses[courseIndex], course)
            this.courses.courses = courses
        }
        return this.grpcSend<Void>(new Void())
    }

    public getCourse(courseID: bigint): Promise<IGrpcResponse<Course>> {
        const course = this.courses.courses.find(c => c.ID === courseID)
        return this.grpcSend<Course>(course)
    }

    public getCourses(): Promise<IGrpcResponse<Courses>> {
        return this.grpcSend<Courses>(this.courses)
    }

    public updateCourseVisibility(request: Enrollment): Promise<IGrpcResponse<Void>> {
        if (this.currentUser === null) {
            return this.grpcSend<Void>(new Void())
        }
        const index = this.enrollments.enrollments.findIndex(e => e.userID === this.currentUser?.ID)
        if (index > -1) {
            const enrollments = this.enrollments.enrollments
            enrollments[index].state = request.state
            this.enrollments.enrollments = enrollments
        }
        return this.grpcSend<Void>(new Void())
    }

    // /* ASSIGNMENTS */ //

    public getAssignments(courseID: bigint): Promise<IGrpcResponse<Assignments>> {
        const assignments = new Assignments()
        for (const assignment of this.assignments.assignments) {
            if (assignment.CourseID === courseID) {
                const benchmarks = this.templateBenchmarks.filter(b => b.AssignmentID === assignment.ID)
                if (benchmarks.length > 0) {
                    assignment.gradingBenchmarks = benchmarks
                }
                assignments.assignments = assignments.assignments.concat(assignment)
            }
        }
        if (assignments.assignments.length === 0) {
            return this.grpcSend<Assignments>(null)
        }
        return this.grpcSend<Assignments>(assignments)
    }

    public updateAssignments(courseID: bigint): Promise<IGrpcResponse<Void>> {
        const course = this.courses.courses.find(c => c.ID === courseID)
        if (!course) {
            return this.grpcSend<Void>(null, new Status({ Error: "Course not found", Code: BigInt(Code.Unknown) }))
        }
        return this.grpcSend<Void>(new Void())
    }

    // /* ENROLLMENTS */ //

    public getEnrollmentsByUser(userID: bigint, statuses?: Enrollment_UserStatus[]): Promise<IGrpcResponse<Enrollments>> {
        if (this.currentUser === null) {
            return this.grpcSend<Enrollments>(null)
        }
        const enrollments = new Enrollments()
        const enrollmentsList: Enrollment[] = []
        this.enrollments.enrollments.forEach(e => {
            const enrollment = e.clone()
            if (enrollment.userID === userID && userID === this.currentUser?.ID && (!statuses || statuses.includes(enrollment.status))) {
                const course = this.courses.courses.find(c => c.ID === enrollment.ID)
                if (course) {
                    enrollment.course = course
                }
                const group = this.groups.groups.find(g => g.ID === enrollment.groupID)
                if (group) {
                    enrollment.group = group
                }
                enrollmentsList.push(enrollment)
            }
        })
        enrollments.enrollments = enrollmentsList
        return this.grpcSend<Enrollments>(enrollments)
    }

    public getEnrollmentsByCourse(courseID: bigint, statuses?: Enrollment_UserStatus[]):
        Promise<IGrpcResponse<Enrollments>> {

        const enrollmentList = this.enrollments.enrollments.filter(e => e.courseID === courseID && (!statuses || statuses.length === 0 || statuses.includes(e.status)))
        if (enrollmentList.length === 0) {
            return this.grpcSend<Enrollments>(null)
        }
        enrollmentList.forEach(e => {
            e.user = this.users.users.find(u => u.ID === e.userID)
        })
        const enrollments = new Enrollments({ enrollments: enrollmentList })
        return this.grpcSend<Enrollments>(enrollments)
    }

    public createEnrollment(courseID: bigint, userID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new Enrollment()
        request.ID = this.generateID(Generate.Enrollment)
        request.userID = userID
        request.courseID = courseID
        const course = this.courses.courses.find(c => c.ID === courseID)
        if (course) {
            request.course = course
            request.status = Enrollment_UserStatus.PENDING
        }
        if (!this.enrollments.enrollments.find(e => e.userID === userID && e.courseID === courseID)) {
            this.enrollments.enrollments = (this.enrollments.enrollments.concat(request))
        }
        return this.grpcSend<Void>(new Void())
    }

    public updateEnrollments(enrollments: Enrollment[]): Promise<IGrpcResponse<Void>> {
        this.enrollments.enrollments.forEach((e, i) => {
            const enrollment = enrollments.find(en => en.ID === e.ID && en.courseID === e.courseID)
            if (enrollment) {
                this.enrollments.enrollments[i].status = (enrollment.status)
            }
        })
        return this.grpcSend<Void>(new Void(), new Status())
    }

    // /* GROUPS */ //

    public getGroup(groupID: bigint): Promise<IGrpcResponse<Group>> {
        return this.grpcSend<Group>(this.groups.groups.find(g => g.ID === groupID))
    }

    public getGroupByUserAndCourse(courseID: bigint, userID: bigint): Promise<IGrpcResponse<Group>> {
        // TODO: Check this
        const group = this.groups.groups.find(g => g.courseID === courseID && g.users.find(u => u.ID === userID))
        if (!group) {
            return this.grpcSend<Group>(null)
        }
        return this.grpcSend<Group>(group)
    }

    public getGroupsByCourse(courseID: bigint): Promise<IGrpcResponse<Groups>> {
        const groups = this.groups.groups.filter(g => g.courseID === courseID)
        if (groups.length === 0) {
            return this.grpcSend<Groups>(null)
        }
        groups.forEach(group => {
            const groupEnrollments = this.enrollments.enrollments.filter(e => e.groupID === group.ID)
            group.enrollments = (groupEnrollments)
            const users: User[] = []
            groupEnrollments.forEach(e => {
                const user = this.users.users.find(u => u.ID === e.userID)
                if (user) {
                    users.push(user)
                }
            })
            group.users = (users)
        })
        return this.grpcSend<Groups>(new Groups({ groups }))
    }

    public updateGroupStatus(groupID: bigint, status: Group_GroupStatus): Promise<IGrpcResponse<Void>> {
        const group = this.groups.groups.findIndex(g => g.ID === groupID)
        if (group > 0) {
            this.groups.groups[group].status = (status)
        }
        return this.grpcSend<Void>(new Void())
    }

    public updateGroup(group: Group): Promise<IGrpcResponse<Group>> {
        const groupID = group.ID
        const currentGroup = this.groups.groups.find(g => g.ID === groupID && g.courseID === group.courseID)
        if (!currentGroup) {
            return this.grpcSend<Group>(new Void(), new Status({ Code: BigInt(Code.NotFound) }))
        }
        // Remove enrollments where the user is not in the group
        const updatedUsers = group.users.map(u => u.ID)
        const currentUsers = currentGroup.users.map(u => u.ID)

        // Merge current and updated users, without duplicates
        const combinedUsers = Array.from(new Set([...updatedUsers, ...currentUsers]))

        combinedUsers.forEach(user => {
            if (!updatedUsers.includes(user)) {
                // Remove user from group
                combinedUsers.splice(combinedUsers.indexOf(user), 1)

                // Unset group ID for enrollment
                this.enrollments.enrollments.forEach(e => {
                    if (e.groupID === groupID && e.userID === user && e.courseID === group.courseID) {
                        e.groupID = BigInt(0)
                    }
                })
            }

            if (!currentUsers.includes(user)) {
                // Add group ID to enrollment, if an enrollment exists for the user
                this.enrollments.enrollments.forEach(e => {
                    if (e.userID === user && e.courseID === group.courseID) {
                        e.groupID = groupID
                    }
                })
            }
        })

        // Update group users and enrollments
        const updatedEnrollments = this.enrollments.enrollments.filter(e => e.groupID === groupID && combinedUsers.includes(e.userID))
        group.enrollments = (updatedEnrollments)
        group.users = (this.users.users.filter(u => combinedUsers.includes(u.ID)))
        Object.assign(currentGroup, group)

        return this.grpcSend<Group>(group)
    }

    public deleteGroup(courseID: bigint, groupID: bigint): Promise<IGrpcResponse<Void>> {
        const group = this.groups.groups.findIndex(g => g.ID === groupID)
        if (group > 0) {
            this.enrollments.enrollments.forEach(e => {
                if (e.groupID === groupID && e.courseID === courseID) {
                    e.groupID = BigInt(0)
                }
            })
            this.groups.groups.splice(group, 1)
        }
        return this.grpcSend<Void>(new Void())
    }

    public createGroup(courseID: bigint, name: string, users: bigint[]): Promise<IGrpcResponse<Group>> {
        // Check that the group doesn't exist
        const group = this.groups.groups.find(g => g.name === name && g.courseID === courseID)
        if (group) {
            return this.grpcSend<Group>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Group already exists" }))
        }
        const request = new Group()
        request.name = name
        request.courseID = courseID
        request.ID = this.generateID(Generate.Group)
        const groupUsers: User[] = []
        users.forEach((ele) => {
            const user = this.users.users.find(u => u.ID === ele)
            if (user) {
                groupUsers.push(user)
                const enrollment = this.enrollments.enrollments.find(e => e.userID === ele && e.courseID === courseID)
                if (enrollment) {
                    enrollment.groupID = BigInt(request.ID)
                }
            } else {
                return this.grpcSend<Group>(null, new Status({ Error: "User not found", Code: BigInt(Code.Unknown) }))
            }
        })
        if (groupUsers.length > 0) {
            request.users = (groupUsers)
        }
        this.groups.groups.push(request)
        return this.grpcSend<Group>(request)
    }

    // /* SUBMISSIONS */ //

    public getSubmission(courseID: bigint, submissionID: bigint): Promise<IGrpcResponse<Submission>> {
        const enrollment = this.enrollments.enrollments.find(enrollment =>
            enrollment.courseID === courseID &&
            enrollment.userID === this.currentUser?.ID
        )
        if (!enrollment) {
            // Current user is not enrolled in course, not allowed to fetch submissions
            return this.grpcSend<Submission>(null, new Status({ Error: "Not found", Code: BigInt(Code.NotFound) }))
        }
        const submission = this.submissions.submissions.find(s => s.ID === submissionID)
        return this.grpcSend<Submission>(submission)
    }

    public getSubmissions(courseID: bigint, userID: bigint): Promise<IGrpcResponse<Submissions>> {
        // Get all assignment IDs
        const assignmentIDs = this.assignments.assignments.filter(a => a.CourseID === courseID && !a.isGroupLab).map(a => a.ID)
        const submissions = this.submissions.submissions.filter(s => s.userID === userID && assignmentIDs.includes(s.AssignmentID))
        if (submissions.length === 0) {
            return this.grpcSend<Submissions>(null, new Status({ Code: BigInt(Code.Unknown), Error: "No submissions found" }))
        }
        return this.grpcSend<Submissions>(new Submissions({ submissions }))
    }

    public getGroupSubmissions(courseID: bigint, groupID: bigint): Promise<IGrpcResponse<Submissions>> {
        const assignmentIDs = this.assignments.assignments.filter(a => a.CourseID === courseID && a.isGroupLab).map(a => a.ID)
        const submissions = this.submissions.submissions.filter(s => s.groupID === groupID && assignmentIDs.includes(s.AssignmentID))
        if (submissions.length === 0) {
            return this.grpcSend<Submissions>(null, new Status({ Code: BigInt(Code.Unknown), Error: "No submissions found" }))
        }
        return this.grpcSend<Submissions>(new Submissions({ submissions }))
    }

    public getSubmissionsByCourse(courseID: bigint, type: SubmissionRequest_SubmissionType): Promise<IGrpcResponse<CourseSubmissions>> {
        // TODO: Remove `.clone()` when done migrating to AsObject in state
        const users = this.users.users
        const groups = this.groups.groups
        const submissions = new CourseSubmissions()
        const enrollmentLinks: EnrollmentLink[] = []
        const course = this.courses.courses.find(c => c.ID === courseID)
        if (!course) {
            return this.grpcSend<CourseSubmissions>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Course not found" }))
        }
        submissions.course = course.clone()

        const enrollments = this.enrollments.enrollments.filter(e => e.courseID === courseID)
        const aIDs = this.assignments.assignments.filter(a => a.CourseID === courseID).map(a => a.ID)
        enrollments.forEach(enrollment => {
            const link = new EnrollmentLink()
            const enroll = enrollment.clone()
            enroll.user = users.find(u => u.ID === enrollment.userID)?.clone()
            enroll.group = groups.find(g => g.ID === enrollment.groupID)?.clone()
            link.enrollment = enroll
            const subs: SubmissionLink[] = []

            this.assignments.assignments.forEach(assignment => {
                if (!aIDs.includes(assignment.ID)) {
                    return
                }
                const subLink = new SubmissionLink()
                subLink.assignment = assignment.clone()
                let submission: Submission | undefined
                switch (type) {
                    case SubmissionRequest_SubmissionType.ALL:
                        submission = this.submissions.submissions.find(s => s.AssignmentID === assignment.ID && (s.userID === enrollment.userID || (s.groupID > 0 && s.groupID === enrollment.groupID)))
                        break
                    case SubmissionRequest_SubmissionType.USER:
                        submission = this.submissions.submissions.find(s => s.AssignmentID === assignment.ID && s.userID === enrollment.userID)
                        break
                    case SubmissionRequest_SubmissionType.GROUP:
                        submission = this.submissions.submissions.find(s => s.AssignmentID === assignment.ID && s.groupID > 0 && s.groupID === enrollment.groupID)
                        break
                }

                if (!submission) {
                    subs.push(subLink)
                    return
                }

                subLink.submission = submission.clone()
                subs.push(subLink)
            })
            link.submissions = subs
            enrollmentLinks.push(link)
        })
        submissions.links = enrollmentLinks
        // TODO
        return this.grpcSend<CourseSubmissions>(submissions)
    }

    public updateSubmission(courseID: bigint, submission: Submission): Promise<IGrpcResponse<Void>> {
        if (!this.courses.courses.find(c => c.ID === courseID)) {
            return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Course not found" }))
        }
        const sub = this.submissions.submissions.find(s => s.ID === submission.ID)
        if (sub) {
            Object.assign(sub, submission)
        }
        return this.grpcSend<Void>(new Void())
    }

    public updateSubmissions(assignmentID: bigint, courseID: bigint, score: number, release: boolean, approve: boolean): Promise<IGrpcResponse<Void>> {
        const assignment = this.assignments.assignments.find(assignment => assignment.ID === assignmentID && assignment.CourseID === courseID)
        if (!assignment) {
            return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Assignment not found" }))
        }

        for (const submission of this.submissions.submissions) {
            if (submission.AssignmentID !== assignmentID) {
                continue
            }
            if (submission.score < score) {
                continue
            }
            if (approve) {
                submission.status = (Submission_Status.APPROVED)
            }
            if (release) {
                submission.released = release
            }
        }
        return this.grpcSend<Void>(new Void())
    }

    public rebuildSubmission(assignmentID: bigint, submissionID: bigint): Promise<IGrpcResponse<Void>> {
        if (this.submissions.submissions.find(sub => sub.ID === submissionID && sub.AssignmentID === assignmentID)) {
            return this.grpcSend<Void>(new Void())
        }
        return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Submission not found" }))
    }

    public rebuildSubmissions(assignmentID: bigint, courseID: bigint): Promise<IGrpcResponse<Void>> {
        if (this.assignments.assignments.find(ass => ass.ID === assignmentID && ass.CourseID === courseID)) {
            return this.grpcSend<Void>(new Void())
        }
        return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Assignment not found" }))
    }

    // /* MANUAL GRADING */ //

    // TODO: All manual grading functions
    public createBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<GradingBenchmark>> {
        bm.ID = (this.generateID(Generate.TemplateBenchmark))
        this.templateBenchmarks.push(bm)
        return this.grpcSend<GradingBenchmark>(bm)
    }

    public createCriterion(c: GradingCriterion): Promise<IGrpcResponse<GradingCriterion>> {
        const benchmarks = this.templateBenchmarks.find(bm => bm.ID === c.BenchmarkID)
        if (!benchmarks) {
            return this.grpcSend<GradingCriterion>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Benchmark not found" }))
        }
        c.ID = (this.generateID(Generate.TemplateCriterion))
        benchmarks.criteria.push(c)
        return this.grpcSend<GradingCriterion>(c)
    }

    public updateBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        const foundIdx = this.templateBenchmarks.findIndex(b => b.ID === bm.ID)
        if (foundIdx === -1) {
            return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Benchmark not found" }))
        }
        Object.assign(this.templateBenchmarks[foundIdx], bm)
        return this.grpcSend<Void>(bm)
    }

    public updateCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        this.templateBenchmarks.forEach(bm => {
            if (bm.ID !== c.BenchmarkID) {
                return
            }
            const index = bm.criteria.findIndex(cr => cr.ID === c.ID)
            if (index !== -1) {
                Object.assign(bm.criteria[index], c)
            }
        })

        return this.grpcSend<Void>(c)
    }

    public deleteBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(bm)
    }

    public deleteCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(c)
    }

    public createReview(r: Review, courseID: bigint): Promise<IGrpcResponse<Review>> {
        if (this.courses.courses.find(c => c.ID === courseID)) {
            return this.grpcSend<Review>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Course not found" }))
        }
        const submission = this.submissions.submissions.find(s => s.ID === r.SubmissionID)
        if (!submission) {
            return this.grpcSend<Review>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Submission not found" }))
        }
        const review = new Review()
        review.ReviewerID = r.ReviewerID
        review.SubmissionID = r.SubmissionID
        review.ID = this.generateID(Generate.Review)

        const benchmarks = this.templateBenchmarks.filter(bm =>
            bm.AssignmentID === submission.AssignmentID
        )
        review.gradingBenchmarks = benchmarks
        review.edited = new Date().getTime().toString()
        submission.reviews = submission.reviews.concat([review])
        return this.grpcSend<Review>(review)
    }

    public updateReview(r: Review, courseID: bigint): Promise<IGrpcResponse<Review>> {
        if (!this.courses.courses.find(c => c.ID === courseID)) {
            return this.grpcSend<Review>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Course not found" }))
        }
        const submission = this.submissions.submissions.find(s => s.ID === r.SubmissionID)
        if (!submission) {
            return this.grpcSend<Review>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Submission not found" }))
        }
        r.score = this.computeScore(r)
        r.edited = new Date().getTime().toString()
        submission.reviews = (submission.reviews.map(rev => {
            if (rev.ID === r.ID) {
                // Return the updated review
                return r
            }
            // Return the original review
            return rev
        }))
        return this.grpcSend<Review>(r)
    }

    // /* REPOSITORY */ //

    public getRepositories(courseID: bigint, types: Repository_Type[]): Promise<IGrpcResponse<Repositories>> {
        // TODO
        if (!this.courses.courses.find(c => c.ID === courseID)) {
            return this.grpcSend<Repositories>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Course not found" }))
        }
        types.forEach(() => {
            // TODO
        })
        //const repos = this.repositories.getRepositoriesList().filter(r => r.courseID === courseID && types.includes(r.getType()))
        return this.grpcSend<Repositories>(new Repositories())
    }

    // /* ORGANIZATIONS */ //

    public async getOrganization(orgName: string): Promise<IGrpcResponse<Organization>> {
        const org = this.organizations.find(o => o.name === orgName)
        await delay(2000)
        if (!org) {
            return this.grpcSend<Organization>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Organization not found" }))
        }
        return this.grpcSend<Organization>(org)
    }

    public isEmptyRepo(courseID: bigint, userID: bigint, groupID: bigint): Promise<IGrpcResponse<Void>> {
        if (courseID <= 0 || userID <= 0 || groupID <= 0) {
            return this.grpcSend<Void>(null, new Status({ Code: BigInt(Code.Unknown), Error: "Invalid arguments" }))
        }
        return this.grpcSend<Void>(true)
    }

    private grpcSend<T>(data: any, status?: Status): Promise<IGrpcResponse<T>> {
        const grpcPromise = new Promise<IGrpcResponse<T>>((resolve) => {
            const temp: IGrpcResponse<T> = {
                data: data as T,
                status: status ?? new Status,
            }
            resolve(temp)
        })
        return grpcPromise
    }

    private initUsers(): void {
        this.users = new Users()
        const userList: User[] = []
        userList.push(
            new User({
                ID: BigInt(1),
                Name: "Test Testersen",
                Email: "test@testersen.no",
                Login: "Test User",
                StudentID: "9999",
                IsAdmin: true,
            })
        )

        userList.push(
            new User({
                ID: BigInt(2),
                Name: "Admin Admin",
                Email: "admin@admin",
                Login: "Admin",
                StudentID: "1000",
                IsAdmin: true,
            })
        )

        userList.push(
            new User({
                ID: BigInt(3),
                Name: "Test Student",
                Email: "test@student.no",
                Login: "Student",
                AvatarURL: "https://avatars0.githubusercontent.com/u/1?v=4",
                IsAdmin: false,
            })
        )

        userList.push(
            new User({
                ID: BigInt(4),
                Name: "Bob Bobsen",
                Email: "bob@bobsen.no",
                Login: "Bob",
                StudentID: "1234",
                IsAdmin: true,
            })
        )

        userList.push(
            new User({
                ID: BigInt(5),
                Name: "Petter Pan",
                Email: "petter@pan.no",
                StudentID: "2345",
                IsAdmin: false,
            })
        )
        this.users.users = (userList)
        this.idMap.set(Generate.User, userList.length)
    }

    private initAssignments() {
        this.assignments = new Assignments()
        const ts = new Date(2017, 5, 25)
        const a0 = new Assignment()
        const a1 = new Assignment()
        const a2 = new Assignment()
        const a3 = new Assignment()
        const a4 = new Assignment()
        const a5 = new Assignment()
        const a6 = new Assignment()
        const a7 = new Assignment()
        const a8 = new Assignment()
        const a9 = new Assignment()
        const a10 = new Assignment()

        a0.ID = BigInt(1)
        a0.CourseID = BigInt(1)
        a0.name = "Lab 1"
        a0.deadline = ts.toDateString()
        a0.scoreLimit = 80
        a0.order = 1

        a1.ID = BigInt(2)
        a1.CourseID = BigInt(1)
        a1.name = ("Lab 2")
        a1.deadline = ts.toDateString()
        a1.scoreLimit = 80
        a1.order = 2

        a2.ID = BigInt(3)
        a2.CourseID = BigInt(1)
        a2.name = "Lab 3"
        a2.reviewers = 1
        a2.deadline = ts.toDateString()
        a2.scoreLimit = 60
        a2.order = 3

        a3.ID = BigInt(4)
        a3.CourseID = BigInt(1)
        a3.name = "Lab 4"
        a3.deadline = ts.toDateString()
        a3.scoreLimit = 75
        a3.order = 4
        a3.isGroupLab = true

        a4.ID = BigInt(5)
        a4.CourseID = BigInt(2)
        a4.name = "Lab 1"
        a4.deadline = ts.toDateString()
        a4.scoreLimit = 90
        a4.order = 1

        a5.ID = BigInt(6)
        a5.CourseID = BigInt(2)
        a5.name = "Lab 2"
        a5.deadline = ts.toDateString()
        a5.scoreLimit = 85
        a5.order = 2

        a6.ID = BigInt(7)
        a6.CourseID = BigInt(2)
        a6.name = "Lab 3"
        a6.deadline = ts.toDateString()
        a6.scoreLimit = 80
        a6.order = 3

        a7.ID = BigInt(8)
        a7.CourseID = BigInt(3)
        a7.name = "Lab 1"
        a7.deadline = ts.toDateString()
        a7.scoreLimit = 90
        a7.order = 1

        a8.ID = BigInt(9)
        a8.CourseID = BigInt(3)
        a8.name = "Lab 2"
        a8.deadline = ts.toDateString()
        a8.scoreLimit = 85
        a8.order = 2

        a9.ID = BigInt(10)
        a9.CourseID = BigInt(4)
        a9.name = "Lab 1"
        a9.deadline = ts.toDateString()
        a9.scoreLimit = 90
        a9.order = 1

        a10.ID = BigInt(11)
        a10.CourseID = BigInt(5)
        a10.name = "Lab 1"
        a10.deadline = ts.toDateString()
        a10.scoreLimit = 90
        a10.order = 1

        const tempAssignments: Assignment[] = [a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10]
        this.assignments.assignments = tempAssignments
        this.idMap.set(Generate.Assignment, tempAssignments.length)
    }

    private initCourses() {
        this.courses = new Courses()
        const course0 = new Course()
        const course1 = new Course()
        const course2 = new Course()
        const course3 = new Course()
        const course4 = new Course()

        course0.ID = BigInt(1)
        course0.name = "Object Oriented Programming"
        course0.code = "DAT100"
        course0.tag = "Spring"
        course0.year = 2017
        course0.organizationID = BigInt(23650610)
        course0.courseCreatorID = BigInt(1)

        course1.ID = BigInt(2)
        course1.name = "Algorithms and Data Structures"
        course1.code = "DAT200"
        course1.tag = "Spring"
        course1.year = 2017
        course1.organizationID = BigInt(23650611)

        course2.ID = BigInt(3)
        course2.name = "Databases"
        course2.code = "DAT220"
        course2.tag = "Spring"
        course2.year = 2017
        course2.organizationID = BigInt(23650612)

        course3.ID = BigInt(4)
        course3.name = "Communication Technology"
        course3.code = "DAT230"
        course3.tag = "Spring"
        course3.year = 2017
        course3.organizationID = BigInt(23650613)

        course4.ID = BigInt(5)
        course4.name = "Operating Systems"
        course4.code = "DAT320"
        course4.tag = "Spring"
        course4.year = 2017
        course4.organizationID = BigInt(23650614)

        const tempCourses: Course[] = [course0, course1, course2, course3, course4]
        this.courses.courses = tempCourses
        this.idMap.set(Generate.Course, tempCourses.length)
    }

    private addLocalCourseStudent() {
        this.enrollments = new Enrollments()
        const localEnrols: Enrollment[] = []
        localEnrols.push(
            new Enrollment({
                ID: BigInt(1),
                courseID: BigInt(1),
                userID: BigInt(1),
                status: Enrollment_UserStatus.TEACHER,
                state: Enrollment_DisplayState.VISIBLE,
                groupID: BigInt(1),
            })
        )

        localEnrols.push(
            new Enrollment({
                ID: BigInt(2),
                courseID: BigInt(2),
                userID: BigInt(1),
                status: Enrollment_UserStatus.TEACHER,
                state: Enrollment_DisplayState.VISIBLE,
            })
        )

        localEnrols.push(
            new Enrollment({
                ID: BigInt(3),
                courseID: BigInt(1),
                userID: BigInt(2),
                status: Enrollment_UserStatus.STUDENT,
                groupID: BigInt(1),
            })
        )

        localEnrols.push(
            new Enrollment({
                ID: BigInt(4),
                courseID: BigInt(2),
                userID: BigInt(2),
                status: Enrollment_UserStatus.PENDING,
            })
        )

        localEnrols.push(
            new Enrollment({
                ID: BigInt(5),
                courseID: BigInt(1),
                userID: BigInt(3),
                status: Enrollment_UserStatus.STUDENT,
                groupID: BigInt(2),
            })
        )

        localEnrols.push(
            new Enrollment({
                ID: BigInt(6),
                courseID: BigInt(1),
                userID: BigInt(4),
                status: Enrollment_UserStatus.STUDENT,
                groupID: BigInt(2),
            })
        )
        this.enrollments.enrollments = (localEnrols)
        this.idMap.set(Generate.Enrollment, localEnrols.length)
    }

    private initOrganizations(): Organization[] {
        const localOrgs: Organization[] = []
        const localOrg = new Organization()
        localOrg.ID = BigInt(23650610)
        localOrg.name = "test"
        localOrg.avatar = "https://avatars2.githubusercontent.com/u/23650610?v=3"
        localOrgs.push(localOrg)
        this.organizations = localOrgs
        return localOrgs
    }

    private addLocalCourseGroups(): void {
        this.groups = new Groups()

        const group1 = new Group({
            ID: BigInt(1),
            name: "Group 1",
            status: Group_GroupStatus.APPROVED,
            courseID: BigInt(1),
        })

        const group2 = new Group({
            ID: BigInt(2),
            name: "Group 2",
            status: Group_GroupStatus.PENDING,
            courseID: BigInt(1),
        })

        this.groups.groups = [group1, group2]
        this.idMap.set(Generate.Group, 2)
    }

    private addLocalLabInfo() {
        this.submissions = new Submissions()
        this.submissions.submissions = [
            new Submission({
                ID: BigInt(1),
                AssignmentID: BigInt(1),
                userID: BigInt(1),
                status: Submission_Status.APPROVED,
                BuildInfo: new BuildInfo({
                    ID: BigInt(1),
                    SubmissionID: BigInt(1),
                    ExecTime: BigInt(1),
                    BuildDate: new Date(2017, 6, 4).toISOString(),
                    BuildLog: "Build log for submission 1",
                }),
                score: 100,
                commitHash: "abc",
                Scores: [
                    new Score({
                        ID: BigInt(1),
                        SubmissionID: BigInt(1),
                        MaxScore: 10,
                        Score: 10,
                        TestName: "Test 1",
                        Weight: 2
                    }),
                    new Score({
                        ID: BigInt(2),
                        SubmissionID: BigInt(1),
                        MaxScore: 10,
                        Score: 10,
                        TestName: "Test 2",
                        Weight: 2
                    }),
                ],
            }),

            new Submission({
                ID: BigInt(2),
                AssignmentID: BigInt(2),
                userID: BigInt(2),
                score: 75,
                commitHash: "bcd",
            }),

            new Submission({
                ID: BigInt(3),
                AssignmentID: BigInt(3),
                userID: BigInt(1),
                score: 80,
                released: true,
                reviews: [
                    new Review({
                        ID: BigInt(1),
                        SubmissionID: BigInt(3),
                        score: 80,
                        feedback: "Well done!",
                        ReviewerID: BigInt(1),
                        gradingBenchmarks: [
                            new GradingBenchmark({
                                ID: BigInt(1),
                                AssignmentID: BigInt(2),
                                heading: "HTML",
                                ReviewID: BigInt(1),
                                criteria: [
                                    new GradingCriterion({
                                        ID: BigInt(1),
                                        BenchmarkID: BigInt(1),
                                        description: "Add div",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    }),
                                    new GradingCriterion({
                                        ID: BigInt(2),
                                        BenchmarkID: BigInt(1),
                                        description: "Div has text",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    })
                                ]
                            }),
                            new GradingBenchmark({
                                ID: BigInt(2),
                                AssignmentID: BigInt(2),
                                heading: "CSS",
                                ReviewID: BigInt(1),
                                criteria: [
                                    new GradingCriterion({
                                        ID: BigInt(3),
                                        BenchmarkID: BigInt(2),
                                        description: "Div centered",
                                        comment: "Good job!",
                                        grade: GradingCriterion_Grade.PASSED,
                                        points: BigInt(10),
                                    }),
                                    new GradingCriterion({
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
            new Submission({
                ID: BigInt(4),
                AssignmentID: BigInt(3),
                groupID: BigInt(1),
                score: 90,
                commitHash: "def",
            }),
            new Submission({
                ID: BigInt(5),
                AssignmentID: BigInt(5),
                userID: BigInt(1),
                score: 100,
                commitHash: "efg",
            }),

            new Submission({
                ID: BigInt(6),
                AssignmentID: BigInt(1),
                userID: BigInt(3),
                score: 50,
                commitHash: "test",
                status: Submission_Status.NONE,
                BuildInfo: new BuildInfo({
                    ID: BigInt(3),
                    BuildDate: new Date(2022, 6, 4).toISOString(),
                    BuildLog: "Build log for test student",
                    ExecTime: BigInt(1),
                }),
                Scores: [
                    new Score({
                        ID: BigInt(3),
                        MaxScore: 10,
                        Score: 5,
                        SubmissionID: BigInt(6),
                        TestName: "Test 1",
                        TestDetails: "Test details for test 1",
                        Weight: 5,
                    }),
                    new Score({
                        ID: BigInt(4),
                        MaxScore: 10,
                        Score: 7,
                        SubmissionID: BigInt(6),
                        TestName: "Test 2",
                        TestDetails: "Test details for test 2",
                        Weight: 2,
                    }),
                ]
            })

        ]
        this.idMap.set(Generate.Submission, 6)
        this.idMap.set(Generate.Review, 1)
        this.idMap.set(Generate.Score, 4)
        this.idMap.set(Generate.BuildInfo, 3)
        this.idMap.set(Generate.GradingBenchmark, 2)
        this.idMap.set(Generate.GradingCriterion, 4)
    }

    private initBenchmarks() {
        this.templateBenchmarks = []

        this.templateBenchmarks.push(
            new GradingBenchmark({
                ID: BigInt(1),
                AssignmentID: BigInt(1),
                heading: "HTML",
                criteria: [
                    new GradingCriterion({
                        ID: BigInt(1),
                        BenchmarkID: BigInt(1),
                        description: "Add div",
                        points: BigInt(10),
                    }),
                    new GradingCriterion({
                        ID: BigInt(2),
                        BenchmarkID: BigInt(1),
                        description: "Div has text",
                        points: BigInt(10),
                    }),
                ]
            }),
            new GradingBenchmark({
                ID: BigInt(2),
                AssignmentID: BigInt(2),
                heading: "CSS",
                criteria: [
                    new GradingCriterion({
                        ID: BigInt(3),
                        BenchmarkID: BigInt(2),
                        description: "Div centered",
                        points: BigInt(10),
                    }),
                    new GradingCriterion({
                        ID: BigInt(4),
                        BenchmarkID: BigInt(2),
                        description: "Div colored",
                        points: BigInt(10),
                    }),
                ]
            })
        )
        this.idMap.set(Generate.TemplateBenchmark, 2)
        this.idMap.set(Generate.TemplateCriterion, 4)
    }

    private computeScore(r: Review) {
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

    public generateID(key: Generate): bigint {
        const sKey = key.toString()
        const id = this.idMap.get(sKey)
        if (!id) {
            this.idMap.set(sKey, 1)
            return BigInt(1)
        }
        this.idMap.set(sKey, id + 1)
        return BigInt(id + 1)
    }
}
