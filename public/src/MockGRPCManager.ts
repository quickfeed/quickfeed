import {
    Assignments,
    AuthorizationResponse,
    Course,
    CourseSubmissions,
    Courses,
    Enrollment,
    Enrollments,
    GradingBenchmark,
    GradingCriterion,
    Group,
    Groups,
    Organization,
    Providers,
    Repositories,
    Repository,
    Review,
    Status,
    SubmissionsForCourseRequest,
    Submission,
    Submissions,
    SubmissionReviewersRequest,
    User,
    Users,
    Void,
    Reviewers,
    Assignment,
    Organizations,
    EnrollmentLink,
    SubmissionLink,
} from "../proto/qf/qf_pb"
import { delay } from "./Helpers"
import { BuildInfo, Score } from "../proto/kit/score/score_pb"
import { StatusCode } from "grpc-web"

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
        this.initProviders()
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


    private providers: Providers
    private groups: Groups
    private users: Users
    private enrollments: Enrollments
    private currentUser: User | null
    private assignments: Assignments
    private courses: Courses
    private organizations: Organizations
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
        const user = this.users.getUsersList().find(u => u.getId() === id)
        if (user) {
            this.currentUser = user
        } else {
            this.currentUser = null
        }
    }

    public async getUser(): Promise<IGrpcResponse<User>> {
        //await delay(10000)
        return this.grpcSend<User>(this.currentUser)
    }

    public getUsers(): Promise<IGrpcResponse<Users>> {
        if (this.currentUser?.getIsadmin()) {
            return this.grpcSend<Users>(this.users)
        }
        return this.grpcSend<Users>(null)
    }

    public updateUser(user: User): Promise<IGrpcResponse<Void>> {
        if (!this.currentUser?.getIsadmin()) {
            return this.grpcSend<Void>(null, new Status().setCode(StatusCode.UNAUTHENTICATED))
        }
        const usr = this.users.getUsersList()?.findIndex(u => u.getId() === user.getId())
        if (usr > -1) {
            Object.assign(this.users.getUsersList()[usr], user)
        }
        return this.grpcSend<Void>(new Void())
    }

    public isAuthorizedTeacher(): Promise<IGrpcResponse<AuthorizationResponse>> {
        return this.grpcSend<AuthorizationResponse>(new AuthorizationResponse().setIsauthorized(true))
    }

    // /* COURSES */ //

    public createCourse(course: Course): Promise<IGrpcResponse<Course>> {
        let data: Course | null = null
        const found = this.courses.getCoursesList().find(c => c.getId() === course.getId())
        const isAdmin = this.currentUser?.getIsadmin()
        const user = this.currentUser
        if (!found && user && isAdmin) {
            course.setId(this.generateID(Generate.Course))
            course.setCoursecreatorid(user.getId())

            this.courses.getCoursesList().push(course)

            // Create new enrollment
            const enrollment = new Enrollment()
            enrollment.setCourseid(course.getId())
            enrollment.setUserid(user.getId())
            enrollment.setStatus(Enrollment.UserStatus.TEACHER)
            enrollment.setId(this.generateID(Generate.Enrollment))
            enrollment.setCourse(course)
            enrollment.setUser(user)
            enrollment.setSlipdaysremaining(course.getSlipdays())
            this.enrollments.getEnrollmentsList().push(enrollment)

            data = course
        }
        return this.grpcSend<Course>(data)
    }

    public updateCourse(course: Course): Promise<IGrpcResponse<Void>> {
        const courseID = course.getId()
        const c = this.courses.getCoursesList().findIndex(c => c.getId() === courseID)
        if (c > -1) {
            const courses = this.courses.getCoursesList()
            Object.assign(courses[c], course)
            this.courses.setCoursesList(courses)
        }
        return this.grpcSend<Void>(new Void())
    }

    public getCourse(courseID: number): Promise<IGrpcResponse<Course>> {
        const course = this.courses.getCoursesList().find(c => c.getId() === courseID)
        return this.grpcSend<Course>(course)
    }

    public getCourses(): Promise<IGrpcResponse<Courses>> {
        return this.grpcSend<Courses>(this.courses)
    }

    public getCoursesByUser(userID: number, statuses: Enrollment.UserStatus[]): Promise<IGrpcResponse<Courses>> {
        const courses = new Courses()
        const courseList: Course[] = []
        for (const enrollment of this.enrollments.getEnrollmentsList()) {
            if (enrollment.getUserid() === userID && statuses.includes(enrollment.getStatus())) {
                const course = this.courses.getCoursesList().find(c => c.getId() === enrollment.getCourseid())
                if (course) {
                    courseList.push(course)
                }
            }
        }
        return this.grpcSend<Courses>(courses)
    }

    public updateCourseVisibility(request: Enrollment): Promise<IGrpcResponse<Void>> {
        if (this.currentUser === null) {
            return this.grpcSend<Void>(new Void())
        }
        const index = this.enrollments.getEnrollmentsList().findIndex(e => e.getUserid() === this.currentUser?.getId())
        if (index > -1) {
            const enrollments = this.enrollments.getEnrollmentsList()
            enrollments[index].setState(request.getState())
            this.enrollments.setEnrollmentsList(enrollments)
        }
        return this.grpcSend<Void>(new Void())
    }

    // /* ASSIGNMENTS */ //

    public getAssignments(courseID: number): Promise<IGrpcResponse<Assignments>> {
        const assignments = new Assignments()
        for (const assignment of this.assignments.getAssignmentsList()) {
            if (assignment.getCourseid() === courseID) {
                const benchmarks = this.templateBenchmarks.filter(b => b.getAssignmentid() === assignment.getId())
                if (benchmarks.length > 0) {
                    assignment.setGradingbenchmarksList(benchmarks)
                }
                assignments.setAssignmentsList(assignments.getAssignmentsList().concat(assignment))
            }
        }
        if (assignments.getAssignmentsList().length === 0) {
            return this.grpcSend<Assignments>(null)
        }
        return this.grpcSend<Assignments>(assignments)
    }

    public updateAssignments(courseID: number): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(new Void())
    }

    // /* ENROLLMENTS */ //

    public getEnrollmentsByUser(userID: number, statuses?: Enrollment.UserStatus[]): Promise<IGrpcResponse<Enrollments>> {
        if (this.currentUser === null) {
            return this.grpcSend<Enrollments>(null)
        }
        const enrollments = new Enrollments()
        const enrollmentsList: Enrollment[] = []
        this.enrollments.getEnrollmentsList().forEach(e => {
            const enrollment = e.clone()
            if (enrollment.getUserid() === userID && userID === this.currentUser?.getId() && (!statuses || statuses.includes(enrollment.getStatus()))) {
                const course = this.courses.getCoursesList().find(c => c.getId() === enrollment.getCourseid())
                if (course) {
                    enrollment.setCourse(course)
                }
                const group = this.groups.getGroupsList().find(g => g.getId() === enrollment.getGroupid())
                if (group) {
                    enrollment.setGroup(group)
                }
                enrollmentsList.push(enrollment)
            }
        })
        return this.grpcSend<Enrollments>(enrollments.setEnrollmentsList(enrollmentsList))
    }

    public getEnrollmentsByCourse(courseID: number, withoutGroupMembers?: boolean, withActivity?: boolean, statuses?: Enrollment.UserStatus[]):
        Promise<IGrpcResponse<Enrollments>> {

        const enrollmentList = this.enrollments.getEnrollmentsList().filter(e => e.getCourseid() === courseID && (!statuses || statuses.length == 0 || statuses.includes(e.getStatus())))
        if (enrollmentList.length === 0) {
            return this.grpcSend<Enrollments>(null)
        }
        enrollmentList.forEach(e => {
            e.setUser(this.users.getUsersList().find(u => u.getId() === e.getUserid()))
        })
        const enrollments = new Enrollments().setEnrollmentsList(enrollmentList)
        return this.grpcSend<Enrollments>(enrollments)
        // TODO: add group members
        //request.setIgnoregroupmembers(withoutGroupMembers ?? false)
        //request.setWithactivity(withActivity ?? false)
        //request.setStatusesList(statuses ?? [])
    }

    public createEnrollment(courseID: number, userID: number): Promise<IGrpcResponse<Void>> {
        const request = new Enrollment()
        request.setId(this.generateID(Generate.Enrollment))
        request.setUserid(userID)
        request.setCourseid(courseID)
        const course = this.courses.getCoursesList().find(c => c.getId() === courseID)
        if (course) {
            request.setCourse(course)
            request.setStatus(Enrollment.UserStatus.PENDING)
        }
        if (!this.enrollments.getEnrollmentsList().find(e => e.getUserid() === userID && e.getCourseid() === courseID)) {
            this.enrollments.setEnrollmentsList(this.enrollments.getEnrollmentsList().concat(request))
        }
        return this.grpcSend<Void>(new Void())
    }

    public updateEnrollments(enrollments: Enrollment[]): Promise<IGrpcResponse<Void>> {
        this.enrollments.getEnrollmentsList().forEach((e, i) => {
            const enrollment = enrollments.find(en => en.getId() === e.getId() && en.getCourseid() === e.getCourseid())
            if (enrollment) {
                this.enrollments.getEnrollmentsList()[i].setStatus(enrollment.getStatus())
            }
        })
        return this.grpcSend<Void>(new Void(), new Status().setCode(StatusCode.OK))
    }

    // /* GROUPS */ //

    public getGroup(groupID: number): Promise<IGrpcResponse<Group>> {
        return this.grpcSend<Group>(this.groups.getGroupsList().find(g => g.getId() === groupID))
    }

    public getGroupByUserAndCourse(courseID: number, userID: number): Promise<IGrpcResponse<Group>> {
        // TODO: Check this
        const group = this.groups.getGroupsList().find(g => g.getCourseid() === courseID && g.getUsersList().find(u => u.getId() === userID))
        if (!group) {
            return this.grpcSend<Group>(null)
        }
        return this.grpcSend<Group>(group)
    }

    public getGroupsByCourse(courseID: number): Promise<IGrpcResponse<Groups>> {
        const groups = this.groups.getGroupsList().filter(g => g.getCourseid() === courseID)
        if (groups.length === 0) {
            return this.grpcSend<Groups>(null)
        }
        groups.forEach(group => {
            const groupEnrollments = this.enrollments.getEnrollmentsList().filter(e => e.getGroupid() === group.getId())
            group.setEnrollmentsList(groupEnrollments)
            const users: User[] = []
            groupEnrollments.forEach(e => {
                const user = this.users.getUsersList().find(u => u.getId() === e.getUserid())
                if (user) {
                    users.push(user)
                }
            })
            group.setUsersList(users)
        })
        return this.grpcSend<Groups>(new Groups().setGroupsList(groups))
    }

    public updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<IGrpcResponse<Void>> {
        const group = this.groups.getGroupsList().findIndex(g => g.getId() === groupID)
        if (group > 0) {
            this.groups.getGroupsList()[group].setStatus(status)
        }
        return this.grpcSend<Void>(new Void())
    }

    public updateGroup(group: Group): Promise<IGrpcResponse<Group>> {
        const groupID = group.getId()
        const currentGroup = this.groups.getGroupsList().find(g => g.getId() === groupID && g.getCourseid() === group.getCourseid())
        if (currentGroup === undefined) {
            return this.grpcSend<Group>(new Void(), new Status().setCode(StatusCode.NOT_FOUND))
        }
        // Remove enrollments where the user is not in the group
        const updatedUsers = group.getUsersList().map(u => u.getId())
        const currentUsers = currentGroup.getUsersList().map(u => u.getId())

        // Merge current and updated users, without duplicates
        const combinedUsers = Array.from(new Set([...updatedUsers, ...currentUsers]))

        combinedUsers.forEach(user => {
            if (!updatedUsers.includes(user)) {
                // Remove user from group
                combinedUsers.splice(combinedUsers.indexOf(user), 1)

                // Unset group ID for enrollment
                this.enrollments.getEnrollmentsList().forEach(e => {
                    if (e.getGroupid() === groupID && e.getUserid() === user && e.getCourseid() === group.getCourseid()) {
                        e.setGroupid(0)
                    }
                })
            }

            if (!currentUsers.includes(user)) {
                // Add group ID to enrollment, if an enrollment exists for the user
                this.enrollments.getEnrollmentsList().forEach(e => {
                    if (e.getUserid() === user && e.getCourseid() === group.getCourseid()) {
                        e.setGroupid(groupID)
                    }
                })
            }
        })

        // Update group users and enrollments
        const updatedEnrollments = this.enrollments.getEnrollmentsList().filter(e => e.getGroupid() === groupID && combinedUsers.includes(e.getUserid()))
        group.setEnrollmentsList(updatedEnrollments)
        group.setUsersList(this.users.getUsersList().filter(u => combinedUsers.includes(u.getId())))
        Object.assign(currentGroup, group)

        return this.grpcSend<Group>(group)
    }

    public deleteGroup(courseID: number, groupID: number): Promise<IGrpcResponse<Void>> {
        const group = this.groups.getGroupsList().findIndex(g => g.getId() === groupID)
        if (group > 0) {
            this.enrollments.getEnrollmentsList().forEach(e => {
                if (e.getGroupid() === groupID && e.getCourseid() === courseID) {
                    e.setGroupid(0)
                }
            })
            this.groups.getGroupsList().splice(group, 1)
        }
        return this.grpcSend<Void>(new Void())
    }

    public createGroup(courseID: number, name: string, users: number[]): Promise<IGrpcResponse<Group>> {
        // Check that the group doesn't exist
        const group = this.groups.getGroupsList().find(g => g.getName() === name && g.getCourseid() === courseID)
        if (group) {
            return this.grpcSend<Group>(null, new Status().setCode(2).setError('Group already exists'))
        }
        const request = new Group()
        request.setName(name)
        request.setCourseid(courseID)
        request.setId(this.generateID(Generate.Group))
        const groupUsers: User[] = []
        users.forEach((ele) => {
            const user = this.users.getUsersList().find(u => u.getId() === ele)
            if (user) {
                groupUsers.push(user)
                const enrollment = this.enrollments.getEnrollmentsList().find(e => e.getUserid() === ele && e.getCourseid() === courseID)
                if (enrollment) {
                    enrollment.setGroupid(request.getId())
                }
            } else {
                return this.grpcSend<Group>(null, new Status().setCode(2).setError('User not found'))
            }
        })
        if (groupUsers.length > 0) {
            request.setUsersList(groupUsers)
        }
        this.groups.getGroupsList().push(request)
        return this.grpcSend<Group>(request)
    }

    // /* SUBMISSIONS */ //
    public getAllSubmissions(courseID: number, userID: number, groupID: number): Promise<IGrpcResponse<Submissions>> {
        const submissions: Submissions = new Submissions()
        // Get all assignment IDs
        const assignmentIDs = this.assignments.getAssignmentsList().filter(a => a.getCourseid() === courseID).map(a => a.getId())
        if (groupID) {
            const subs = this.submissions.getSubmissionsList().filter(s => s.getGroupid() === groupID && assignmentIDs.includes(s.getAssignmentid()))
            submissions.setSubmissionsList(subs)
        }
        if (userID) {
            const subs = this.submissions.getSubmissionsList().filter(s => s.getUserid() === userID && assignmentIDs.includes(s.getAssignmentid()))
            submissions.setSubmissionsList(subs)
        }
        return this.grpcSend<Submissions>(submissions)
    }

    public getSubmissions(courseID: number, userID: number): Promise<IGrpcResponse<Submissions>> {
        // Get all assignment IDs
        const assignmentIDs = this.assignments.getAssignmentsList().filter(a => a.getCourseid() === courseID && !a.getIsgrouplab()).map(a => a.getId())
        const submissionList = this.submissions.getSubmissionsList().filter(s => s.getUserid() === userID && assignmentIDs.includes(s.getAssignmentid()))
        if (submissionList.length === 0) {
            return this.grpcSend<Submissions>(null, new Status().setCode(2).setError('No submissions found'))
        }
        const submissions = new Submissions().setSubmissionsList(submissionList)
        return this.grpcSend<Submissions>(submissions)
    }

    public getGroupSubmissions(courseID: number, groupID: number): Promise<IGrpcResponse<Submissions>> {
        const assignmentIDs = this.assignments.getAssignmentsList().filter(a => a.getCourseid() === courseID && a.getIsgrouplab()).map(a => a.getId())
        const submissions = this.submissions.getSubmissionsList().filter(s => s.getGroupid() === groupID && assignmentIDs.includes(s.getAssignmentid()))
        if (submissions.length === 0) {
            return this.grpcSend<Submissions>(null, new Status().setCode(2).setError('No submissions found'))
        }
        return this.grpcSend<Submissions>(new Submissions().setSubmissionsList(submissions))
    }

    public getSubmissionsByCourse(courseID: number, type: SubmissionsForCourseRequest.Type, withBuildInfo: boolean): Promise<IGrpcResponse<CourseSubmissions>> {
        // TODO: Remove `.clone()` when done migrating to AsObject in state
        const users = this.users.getUsersList()
        const groups = this.groups.getGroupsList()
        const submissions = new CourseSubmissions()
        const enrollmentLinks: EnrollmentLink[] = []
        const course = this.courses.getCoursesList().find(c => c.getId() === courseID)
        if (!course) {
            return this.grpcSend<CourseSubmissions>(null, new Status().setCode(2).setError('Course not found'))
        }
        submissions.setCourse(course.clone())

        const enrollments = this.enrollments.getEnrollmentsList().filter(e => e.getCourseid() === courseID)
        const aIDs = this.assignments.getAssignmentsList().filter(a => a.getCourseid() === courseID).map(a => a.getId())
        enrollments.forEach(enrollment => {
            const link = new EnrollmentLink()
            const enroll = Object.assign(new Enrollment(), enrollment.clone())
            enroll.setUser(users.find(u => u.getId() === enrollment.getUserid())?.clone())
            enroll.setGroup(groups.find(g => g.getId() === enrollment.getGroupid())?.clone())
            link.setEnrollment(enroll)
            const subs: SubmissionLink[] = []

            this.assignments.getAssignmentsList().forEach(assignment => {
                if (!aIDs.includes(assignment.getId())) {
                    return
                }
                const subLink = new SubmissionLink()
                subLink.setAssignment(assignment.clone())
                let submission: Submission | undefined
                switch (type) {
                    case SubmissionsForCourseRequest.Type.ALL:
                        submission = this.submissions.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId() && (s.getUserid() === enrollment.getUserid() || (s.getGroupid() > 0 && s.getGroupid() === enrollment.getGroupid())))
                        break
                    case SubmissionsForCourseRequest.Type.INDIVIDUAL:
                        submission = this.submissions.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId() && s.getUserid() === enrollment.getUserid())
                        break
                    case SubmissionsForCourseRequest.Type.GROUP:
                        submission = this.submissions.getSubmissionsList().find(s => s.getAssignmentid() === assignment.getId() && s.getGroupid() > 0 && s.getGroupid() === enrollment.getGroupid())
                        break
                }

                if (!submission) {
                    subs.push(subLink)
                    return
                }

                subLink.setSubmission(submission.clone())
                subs.push(subLink)
            })
            link.setSubmissionsList(subs)
            enrollmentLinks.push(link)
        })
        submissions.setLinksList(enrollmentLinks)
        // TODO
        return this.grpcSend<CourseSubmissions>(submissions)
    }

    public updateSubmission(courseID: number, s: Submission): Promise<IGrpcResponse<Void>> {
        const submission = this.submissions.getSubmissionsList().find(s => s.getId() === s.getId())
        if (submission) {
            Object.assign(submission, s)
        }
        return this.grpcSend<Void>(new Void())
    }

    public updateSubmissions(assignmentID: number, courseID: number, score: number, release: boolean, approve: boolean): Promise<IGrpcResponse<Void>> {
        const assignment = this.assignments.getAssignmentsList().find(assignment => assignment.getId() === assignmentID && assignment.getCourseid() === courseID)
        if (!assignment) {
            return this.grpcSend<Void>(null, new Status().setCode(2).setError('Assignment not found'))
        }

        for (const submission of this.submissions.getSubmissionsList()) {
            if (submission.getAssignmentid() !== assignmentID) {
                continue
            }
            if (submission.getScore() < score) {
                continue
            }
            if (approve) {
                submission.setStatus(Submission.Status.APPROVED)
            }
            if (release) {
                submission.setReleased(release)
            }
        }
        return this.grpcSend<Void>(new Void())
    }

    public rebuildSubmission(assignmentID: number, submissionID: number): Promise<IGrpcResponse<Void>> {
        if (this.submissions.getSubmissionsList().find(sub => sub.getId() === submissionID && sub.getAssignmentid() === assignmentID)) {
            return this.grpcSend<Void>(new Void())
        }
        return this.grpcSend<Void>(null, new Status().setCode(2).setError('Submission not found'))
    }

    public rebuildSubmissions(assignmentID: number, courseID: number): Promise<IGrpcResponse<Void>> {
        if (this.assignments.getAssignmentsList().find(ass => ass.getId() === assignmentID && ass.getCourseid() === courseID)) {
            return this.grpcSend<Void>(new Void())
        }
        return this.grpcSend<Void>(null, new Status().setCode(2).setError('Assignment not found'))
    }

    // /* MANUAL GRADING */ //

    // TODO: All manual grading functions
    public createBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<GradingBenchmark>> {
        bm.setId(this.generateID(Generate.TemplateBenchmark))
        this.templateBenchmarks.push(bm)
        return this.grpcSend<GradingBenchmark>(bm)
    }

    public createCriterion(c: GradingCriterion): Promise<IGrpcResponse<GradingCriterion>> {
        const benchmarks = this.templateBenchmarks.find(bm => bm.getId() === c.getBenchmarkid())
        if (!benchmarks) {
            return this.grpcSend<GradingCriterion>(null, new Status().setCode(2).setError('Benchmark not found'))
        }
        c.setId(this.generateID(Generate.TemplateCriterion))
        benchmarks.getCriteriaList().push(c)
        return this.grpcSend<GradingCriterion>(c)
    }

    public updateBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        const foundIdx = this.templateBenchmarks.findIndex(b => b.getId() === bm.getId())
        if (foundIdx === -1) {
            return this.grpcSend<Void>(null, new Status().setCode(2).setError('Benchmark not found'))
        }
        Object.assign(this.templateBenchmarks[foundIdx], bm)
        return this.grpcSend<Void>(bm)
    }

    public updateCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        this.templateBenchmarks.forEach(bm => {
            if (bm.getId() !== c.getBenchmarkid()) {
                return
            }
            const index = bm.getCriteriaList().findIndex(cr => cr.getId() === c.getId())
            if (index !== -1) {
                Object.assign(bm.getCriteriaList()[index], c)
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

    public createReview(r: Review, courseID: number): Promise<IGrpcResponse<Review>> {
        const submission = this.submissions.getSubmissionsList().find(s => s.getId() === r.getSubmissionid())
        if (!submission) {
            return this.grpcSend<Review>(null, new Status().setCode(2).setError('Submission not found'))
        }
        const review = new Review()
        review.setReviewerid(r.getReviewerid())
        review.setSubmissionid(r.getSubmissionid())
        review.setId(this.generateID(Generate.Review))

        const benchmarks = this.templateBenchmarks.filter(bm =>
            bm.getAssignmentid() === submission.getAssignmentid()
        )
        review.setGradingbenchmarksList(benchmarks)
        review.setEdited(new Date().getTime().toString())
        submission.setReviewsList(submission.getReviewsList().concat([review]))
        return this.grpcSend<Review>(review)
    }

    public updateReview(r: Review, courseID: number): Promise<IGrpcResponse<Review>> {
        const submission = this.submissions.getSubmissionsList().find(s => s.getId() === r.getSubmissionid())
        if (!submission) {
            return this.grpcSend<Review>(null, new Status().setCode(2).setError('Submission not found'))
        }
        r.setScore(this.computeScore(r))
        r.setEdited(new Date().getTime().toString())
        submission.setReviewsList(submission.getReviewsList().map(rev => {
            if (rev.getId() === r.getId()) {
                // Return the updated review
                return r
            }
            // Return the original review
            return rev
        }))
        return this.grpcSend<Review>(r)
    }

    public getReviewers(submissionID: number, courseID: number): Promise<IGrpcResponse<Reviewers>> {
        const request = new SubmissionReviewersRequest()
        request.setSubmissionid(submissionID)
        request.setCourseid(courseID)
        return this.grpcSend<Reviewers>(new Reviewers())
    }

    // /* REPOSITORY */ //

    public getRepositories(courseID: number, types: Repository.Type[]): Promise<IGrpcResponse<Repositories>> {
        // TODO
        //const repos = this.repositories.getRepositoriesList().filter(r => r.getCourseid() === courseID && types.includes(r.getType()))
        return this.grpcSend<Repositories>(new Repositories())
    }

    // /* ORGANIZATIONS */ //

    public async getOrganization(orgName: string): Promise<IGrpcResponse<Organization>> {
        const org = this.organizations.getOrganizationsList().find(o => o.getPath() === orgName)
        await delay(2000)
        if (!org) {
            return this.grpcSend<Organization>(null, new Status().setCode(2).setError('Organization not found'))
        }
        return this.grpcSend<Organization>(org)
    }

    public getProviders(): Promise<IGrpcResponse<Providers>> {
        return this.grpcSend<Providers>(this.providers)
    }

    public isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<IGrpcResponse<Void>> {
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

    private initProviders(): void {
        this.providers = new Providers()
        this.providers.setProvidersList([
            "github",
        ])
    }

    private initUsers(): void {
        this.users = new Users()
        const userList: User[] = []
        userList.push(
            new User()
                .setId(1)
                .setName("Test Testersen")
                .setEmail("test@testersen.no")
                .setLogin("Test User")
                .setStudentid("9999")
                .setIsadmin(true)
        )

        userList.push(
            new User()
                .setId(2)
                .setName("Admin Admin")
                .setEmail("admin@admin")
                .setLogin("Admin")
                .setStudentid("1000")
                .setIsadmin(true)
        )

        userList.push(
            new User()
                .setId(3)
                .setName("Test Student")
                .setEmail("test@student.no")
                .setLogin("Student")
                .setAvatarurl("https://avatars0.githubusercontent.com/u/1?v=4")
                .setStudentid("1234")
                .setIsadmin(false)
        )

        userList.push(
            new User()
                .setId(4)
                .setName("Bob Bobsen")
                .setEmail("bob@bobsen.no")
                .setStudentid("1234")
                .setIsadmin(true)
        )

        userList.push(
            new User()
                .setId(5)
                .setName("Petter Pan")
                .setEmail("petter@pan.no")
                .setStudentid("1234")
                .setIsadmin(true)
        )
        this.users.setUsersList(userList)
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

        a0.setId(1)
        a0.setCourseid(1)
        a0.setName("Lab 1")
        a0.setScriptfile("Go")
        a0.setDeadline(ts.toDateString())
        a0.setScorelimit(80)
        a0.setOrder(1)

        a1.setId(2)
        a1.setCourseid(1)
        a1.setName("Lab 2")
        a1.setScriptfile("Go")
        a1.setDeadline(ts.toDateString())
        a1.setScorelimit(80)
        a1.setOrder(2)

        a2.setId(3)
        a2.setCourseid(1)
        a2.setName("Lab 3")
        a2.setReviewers(1)
        a2.setDeadline(ts.toDateString())
        a2.setScorelimit(60)
        a2.setOrder(3)

        a3.setId(4)
        a3.setCourseid(1)
        a3.setName("Lab 4")
        a3.setScriptfile("Go")
        a3.setDeadline(ts.toDateString())
        a3.setScorelimit(75)
        a3.setOrder(4)
        a3.setIsgrouplab(true)

        a4.setId(5)
        a4.setCourseid(2)
        a4.setName("Lab 1")
        a4.setScriptfile("Go")
        a4.setDeadline(ts.toDateString())
        a4.setScorelimit(90)
        a4.setOrder(1)

        a5.setId(6)
        a5.setCourseid(2)
        a5.setName("Lab 2")
        a5.setScriptfile("Go")
        a5.setDeadline(ts.toDateString())
        a5.setScorelimit(85)
        a5.setOrder(2)

        a6.setId(7)
        a6.setCourseid(2)
        a6.setName("Lab 3")
        a6.setScriptfile("Go")
        a6.setDeadline(ts.toDateString())
        a6.setScorelimit(80)
        a6.setOrder(3)

        a7.setId(8)
        a7.setCourseid(3)
        a7.setName("Lab 1")
        a7.setScriptfile("TypeScript")
        a7.setDeadline(ts.toDateString())
        a7.setScorelimit(90)
        a7.setOrder(1)

        a8.setId(9)
        a8.setCourseid(3)
        a8.setName("Lab 2")
        a8.setScriptfile("Go")
        a8.setDeadline(ts.toDateString())
        a8.setScorelimit(85)
        a8.setOrder(2)

        a9.setId(10)
        a9.setCourseid(4)
        a9.setName("Lab 1")
        a9.setScriptfile("Go")
        a9.setDeadline(ts.toDateString())
        a9.setScorelimit(90)
        a9.setOrder(1)

        a10.setId(11)
        a10.setCourseid(5)
        a10.setName("Lab 1")
        a10.setScriptfile("TypeScript")
        a10.setDeadline(ts.toDateString())
        a10.setScorelimit(90)
        a10.setOrder(1)

        const tempAssignments: Assignment[] = [a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10]
        this.assignments.setAssignmentsList(tempAssignments)
        this.idMap.set(Generate.Assignment, tempAssignments.length)
    }

    private initCourses() {
        this.courses = new Courses()
        const course0 = new Course()
        const course1 = new Course()
        const course2 = new Course()
        const course3 = new Course()
        const course4 = new Course()

        course0.setId(1)
        course0.setName("Object Oriented Programming")
        course0.setCode("DAT100")
        course0.setTag("Spring")
        course0.setYear(2017)
        course0.setProvider("github")
        course0.setOrganizationid(23650610)
        course0.setCoursecreatorid(1)

        course1.setId(2)
        course1.setName("Algorithms and Datastructures")
        course1.setCode("DAT200")
        course1.setTag("Spring")
        course1.setYear(2017)
        course1.setProvider("github")
        course1.setOrganizationid(23650611)

        course2.setId(3)
        course2.setName("Databases")
        course2.setCode("DAT220")
        course2.setTag("Spring")
        course2.setYear(2017)
        course2.setProvider("github")
        course2.setOrganizationid(23650612)

        course3.setId(4)
        course3.setName("Communication Technology")
        course3.setCode("DAT230")
        course3.setTag("Spring")
        course3.setYear(2017)
        course3.setProvider("github")
        course3.setOrganizationid(23650613)

        course4.setId(5)
        course4.setName("Operating Systems")
        course4.setCode("DAT320")
        course4.setTag("Spring")
        course4.setYear(2017)
        course4.setProvider("github")
        course4.setOrganizationid(23650614)

        const tempCourses: Course[] = [course0, course1, course2, course3, course4]
        this.courses.setCoursesList(tempCourses)
        this.idMap.set(Generate.Course, tempCourses.length)
    }

    private addLocalCourseStudent() {
        this.enrollments = new Enrollments()
        const localEnrols: Enrollment[] = []
        localEnrols.push(
            new Enrollment()
                .setId(1)
                .setCourseid(1)
                .setUserid(1)
                .setStatus(Enrollment.UserStatus.TEACHER)
                .setState(2)
                .setGroupid(1)
        )

        localEnrols.push(
            new Enrollment()
                .setId(2)
                .setCourseid(2)
                .setUserid(1)
                .setStatus(Enrollment.UserStatus.TEACHER)
                .setState(2)
        )

        localEnrols.push(
            new Enrollment()
                .setId(3)
                .setCourseid(1)
                .setUserid(2)
                .setStatus(Enrollment.UserStatus.STUDENT)
                .setGroupid(1)
        )

        localEnrols.push(
            new Enrollment()
                .setId(4)
                .setCourseid(2)
                .setUserid(2)
                .setStatus(Enrollment.UserStatus.PENDING)
        )

        localEnrols.push(
            new Enrollment()
                .setId(5)
                .setCourseid(1)
                .setUserid(3)
                .setStatus(Enrollment.UserStatus.STUDENT)
                .setGroupid(2)
        )

        localEnrols.push(
            new Enrollment()
                .setId(6)
                .setCourseid(1)
                .setUserid(4)
                .setStatus(Enrollment.UserStatus.STUDENT)
                .setGroupid(2)
        )
        this.enrollments.setEnrollmentsList(localEnrols)
        this.idMap.set(Generate.Enrollment, localEnrols.length)
    }

    private initOrganizations(): Organization[] {
        this.organizations = new Organizations()
        const localOrgs: Organization[] = []
        const localOrg = new Organization()
        localOrg.setId(23650610)
        localOrg.setPath("test")
        localOrg.setAvatar("https://avatars2.githubusercontent.com/u/23650610?v=3")
        localOrgs.push(localOrg)
        this.organizations.setOrganizationsList(localOrgs)
        return localOrgs
    }

    private addLocalCourseGroups(): void {
        this.groups = new Groups()

        const group1 = new Group()
        group1.setId(1)
        group1.setName("Group 1")
        group1.setStatus(Group.GroupStatus.APPROVED)
        group1.setCourseid(1)


        const group2 = new Group()
        group2.setId(2)
        group2.setName("Group 2")
        group2.setStatus(Group.GroupStatus.PENDING)
        group2.setCourseid(1)

        this.groups.setGroupsList([group1, group2])
        this.idMap.set(Generate.Group, 2)
    }

    private addLocalLabInfo() {
        this.submissions = new Submissions()
        this.submissions.setSubmissionsList([
            new Submission()
                .setId(1)
                .setAssignmentid(1)
                .setUserid(1)
                .setStatus(Submission.Status.APPROVED)
                .setBuildinfo(
                    new BuildInfo()
                        .setId(1)
                        .setSubmissionid(1)
                        .setExectime(1)
                        .setBuilddate(new Date(2017, 6, 4).toISOString())
                        .setBuildlog("Build log for build 1")
                )
                .setScore(100)
                .setCommithash("abc")
                .setScoresList([
                    new Score()
                        .setId(1)
                        .setMaxscore(10)
                        .setScore(10)
                        .setTestname("Test 1")
                        .setSubmissionid(1)
                        .setWeight(2),
                    new Score()
                        .setId(2)
                        .setMaxscore(10)
                        .setScore(10)
                        .setTestname("Test 2")
                        .setSubmissionid(1)
                        .setWeight(2),

                ])
            ,
            new Submission()
                .setId(2)
                .setAssignmentid(2)
                .setUserid(2)
                .setScore(75)
                .setCommithash("bcd"),

            new Submission()
                .setId(3)
                .setAssignmentid(3)
                .setUserid(1)
                .setScore(80)
                .setReleased(true)
                .setReviewsList([
                    new Review()
                        .setId(1)
                        .setScore(80)
                        .setSubmissionid(3)
                        .setFeedback("Well done!")
                        .setReviewerid(1)
                        .setGradingbenchmarksList([
                            new GradingBenchmark()
                                .setAssignmentid(2)
                                .setHeading("HTML")
                                .setId(1)
                                .setReviewid(1)
                                .setCriteriaList([
                                    new GradingCriterion()
                                        .setId(1)
                                        .setBenchmarkid(1)
                                        .setDescription("Add div")
                                        .setComment("Good job!")
                                        .setGrade(GradingCriterion.Grade.PASSED)
                                        .setPoints(10),
                                    new GradingCriterion()
                                        .setId(2)
                                        .setBenchmarkid(1)
                                        .setDescription("Div has text")
                                        .setComment("Good job!")
                                        .setGrade(GradingCriterion.Grade.PASSED)
                                        .setPoints(10),
                                ]),
                            new GradingBenchmark()
                                .setAssignmentid(2)
                                .setHeading("CSS")
                                .setId(2)
                                .setReviewid(1)
                                .setCriteriaList([
                                    new GradingCriterion()
                                        .setId(3)
                                        .setBenchmarkid(2)
                                        .setDescription("Div centered")
                                        .setComment("Good job!")
                                        .setGrade(GradingCriterion.Grade.PASSED)
                                        .setPoints(10),
                                    new GradingCriterion()
                                        .setId(4)
                                        .setBenchmarkid(2)
                                        .setDescription("Div colored")
                                        .setComment("Good job!")
                                        .setGrade(GradingCriterion.Grade.PASSED)
                                        .setPoints(10),
                                ])
                        ])
                ]),

            new Submission()
                .setId(4)
                .setAssignmentid(3)
                .setGroupid(1)
                .setScore(90)
                .setCommithash("def"),

            new Submission()
                .setId(5)
                .setAssignmentid(5)
                .setUserid(1)
                .setScore(100)
                .setCommithash("efg"),

            new Submission()
                .setId(6)
                .setAssignmentid(1)
                .setUserid(3)
                .setScore(50)
                .setCommithash("test")
                .setStatus(0)
                .setBuildinfo(
                    new BuildInfo()
                        .setId(3)
                        .setBuilddate(new Date(2022, 6, 4).toISOString())
                        .setBuildlog("Build log for test student")
                        .setExectime(1)
                )
                .setScoresList(
                    [
                        new Score()
                            .setId(3)
                            .setMaxscore(10)
                            .setScore(5)
                            .setSubmissionid(6)
                            .setTestname("Test 1")
                            .setTestdetails("Test details")
                            .setWeight(5),

                        new Score()
                            .setId(4)
                            .setMaxscore(10)
                            .setScore(7)
                            .setTestname("Test 2")
                            .setTestdetails("Test details")
                            .setSubmissionid(6)
                            .setWeight(2),
                    ]
                )
        ]
        )
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
            new GradingBenchmark()
                .setId(1)
                .setAssignmentid(2)
                .setHeading("HTML")
                .setCriteriaList([
                    new GradingCriterion()
                        .setId(1)
                        .setBenchmarkid(1)
                        .setDescription("Add div")
                        .setPoints(10),
                    new GradingCriterion()
                        .setId(2)
                        .setDescription("Div has text")
                        .setPoints(10),
                ]),
            new GradingBenchmark()
                .setId(2)
                .setAssignmentid(2)
                .setHeading("CSS")
                .setCriteriaList([
                    new GradingCriterion()
                        .setId(3)
                        .setDescription("Div centered")
                        .setPoints(10),
                    new GradingCriterion()
                        .setId(4)
                        .setDescription("Div colored")
                        .setPoints(10),
                ])
        )
        this.idMap.set(Generate.TemplateBenchmark, 2)
        this.idMap.set(Generate.TemplateCriterion, 4)
    }

    private computeScore(r: Review) {
        let score = 0
        let totalApproved = 0
        let total = 0
        for (let i = 0; i < r.getGradingbenchmarksList().length; i++) {
            const gb = r.getGradingbenchmarksList()[i]
            for (let j = 0; j < gb.getCriteriaList().length; j++) {
                const criterion = gb.getCriteriaList()[j]
                total++
                if (criterion.getGrade() == GradingCriterion.Grade.PASSED) {
                    score += criterion.getPoints()
                    totalApproved++
                }
            }
        }
        if (score == 0) {
            score = 100 / total * totalApproved
        }
        return score
    }

    public generateID(key: Generate): number {
        const skey = key.toString()
        const id = this.idMap.get(skey)
        if (!id) {
            this.idMap.set(skey, 1)
            return 1
        }
        this.idMap.set(skey, id + 1)
        return id + 1
    }
}
