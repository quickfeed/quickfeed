import {
    Assignments,
    Course,
    Courses,
    Enrollment,
    Enrollments,
    Enrollment_UserStatus,
    GradingBenchmark,
    GradingCriterion,
    Group,
    Groups,
    Group_GroupStatus,
    Repository_Type,
    Review,
    Submission,
    Submissions,
    User,
    Users,
} from "../proto/qf/types_pb"
import {
    CourseRequest,
    CourseSubmissions,
    EnrollmentStatusRequest,
    EnrollmentRequest,
    GetGroupRequest,
    GroupRequest,
    Organization,
    OrgRequest,
    RebuildRequest,
    Repositories,
    RepositoryRequest,
    ReviewRequest,
    Status,
    SubmissionRequest,
    SubmissionsForCourseRequest,
    SubmissionReviewersRequest,
    UpdateSubmissionRequest,
    UpdateSubmissionsRequest,
    URLRequest,
    Void,
    Reviewers,
    SubmissionsForCourseRequest_Type,
} from "../proto/qf/requests_pb"
import { QuickFeedService } from "../proto/qf/quickfeed_connectweb"
import { createConnectTransport, ConnectError, createCallbackClient, CallbackClient } from "@bufbuild/connect-web"

export interface IGrpcResponse<T> {
    status: Status
    data?: T
}

export class GrpcManager {

    private agService: CallbackClient<typeof QuickFeedService>

    constructor() {
        const transport = createConnectTransport({
            baseUrl: "https://" + window.location.host,
        })
        this.agService = createCallbackClient(QuickFeedService, transport)
    }



    public getUser(): Promise<IGrpcResponse<User>> {
        return this.grpcSend<User>(this.agService.getUser, new Void())
    }

    public getUsers(): Promise<IGrpcResponse<Users>> {
        return this.grpcSend<Users>(this.agService.getUsers, new Void())
    }

    public updateUser(user: User): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateUser, user)
    }

    // /* COURSES */ //

    public createCourse(course: Course): Promise<IGrpcResponse<Course>> {
        return this.grpcSend<Course>(this.agService.createCourse, course)
    }

    public updateCourse(course: Course): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateCourse, course)
    }

    public getCourse(courseID: bigint): Promise<IGrpcResponse<Course>> {
        const request = new CourseRequest({ courseID: courseID })
        return this.grpcSend<Course>(this.agService.getCourse, request)
    }

    public getCourses(): Promise<IGrpcResponse<Courses>> {
        return this.grpcSend<Courses>(this.agService.getCourses, new Void())
    }

    public getCoursesByUser(userID: bigint, statuses: Enrollment_UserStatus[]): Promise<IGrpcResponse<Courses>> {
        const request = new EnrollmentStatusRequest({
            userID: userID,
            statuses: statuses,
        })
        return this.grpcSend<Courses>(this.agService.getCoursesByUser, request)
    }

    public updateCourseVisibility(request: Enrollment): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateCourseVisibility, request)
    }

    // /* ASSIGNMENTS */ //

    public getAssignments(courseID: bigint): Promise<IGrpcResponse<Assignments>> {
        const request = new CourseRequest({ courseID: courseID })
        return this.grpcSend<Assignments>(this.agService.getAssignments, request)
    }

    public updateAssignments(courseID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new CourseRequest({ courseID: courseID })
        return this.grpcSend<Void>(this.agService.updateAssignments, request)
    }

    // /* ENROLLMENTS */ //

    public getEnrollmentsByUser(userID: bigint, statuses?: Enrollment_UserStatus[]): Promise<IGrpcResponse<Enrollments>> {
        const request = new EnrollmentStatusRequest({
            userID: userID,
            statuses: statuses,
        })
        return this.grpcSend<Enrollments>(this.agService.getEnrollmentsByUser, request)
    }

    public getEnrollmentsByCourse(courseID: bigint, withoutGroupMembers?: boolean, withActivity?: boolean, statuses?: Enrollment_UserStatus[]):
        Promise<IGrpcResponse<Enrollments>> {
        const request = new EnrollmentRequest({
            courseID: courseID,
            ignoreGroupMembers: withoutGroupMembers ?? false,
            withActivity: withActivity ?? false,
            statuses: statuses,
        })
        return this.grpcSend<Enrollments>(this.agService.getEnrollmentsByCourse, request)
    }

    public createEnrollment(courseID: bigint, userID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new Enrollment({
            courseID: courseID,
            userID: userID,
        })
        return this.grpcSend<Void>(this.agService.createEnrollment, request)
    }

    public updateEnrollments(enrollments: Enrollment[]): Promise<IGrpcResponse<Void>> {
        const request = new Enrollments({
            enrollments: enrollments,
        })
        return this.grpcSend<Void>(this.agService.updateEnrollments, request)
    }

    // /* GROUPS */ //

    public getGroup(groupID: bigint): Promise<IGrpcResponse<Group>> {
        const request = new GetGroupRequest({ groupID: groupID })
        return this.grpcSend<Group>(this.agService.getGroup, request)
    }

    public getGroupByUserAndCourse(courseID: bigint, userID: bigint): Promise<IGrpcResponse<Group>> {
        const request = new GroupRequest({
            courseID: courseID,
            userID: userID,
        })
        return this.grpcSend<Group>(this.agService.getGroupByUserAndCourse, request)
    }

    public getGroupsByCourse(courseID: bigint): Promise<IGrpcResponse<Groups>> {
        const request = new CourseRequest({ courseID: courseID })
        return this.grpcSend<Groups>(this.agService.getGroupsByCourse, request)
    }

    public updateGroupStatus(groupID: bigint, status: Group_GroupStatus): Promise<IGrpcResponse<Void>> {
        const request = new Group({
            ID: groupID,
            status: status,
        })
        return this.grpcSend<Void>(this.agService.updateGroup, request)
    }

    public updateGroup(group: Group): Promise<IGrpcResponse<Group>> {
        return this.grpcSend<Group>(this.agService.updateGroup, group)
    }

    public deleteGroup(courseID: bigint, groupID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new GroupRequest({
            courseID: courseID,
            groupID: groupID,
        })
        return this.grpcSend<Void>(this.agService.deleteGroup, request)
    }

    public createGroup(courseID: bigint, name: string, users: bigint[]): Promise<IGrpcResponse<Group>> {
        const request = new Group({
            courseID: courseID,
            name: name,
            users: users.map(userID => new User({ ID: userID })),
        })
        return this.grpcSend<Group>(this.agService.createGroup, request)
    }

    // /* SUBMISSIONS */ //
    public getAllSubmissions(courseID: bigint, userID: bigint, groupID: bigint): Promise<IGrpcResponse<Submissions>> {
        const request = new SubmissionRequest({
            courseID: courseID,
            userID: userID,
            groupID: groupID,
        })
        return this.grpcSend<Submissions>(this.agService.getSubmissions, request)
    }

    public getSubmissions(courseID: bigint, userID: bigint): Promise<IGrpcResponse<Submissions>> {
        const request = new SubmissionRequest({
            courseID: courseID,
            userID: userID,
        })
        return this.grpcSend<Submissions>(this.agService.getSubmission, request)
    }
    public getSubmission(courseID: bigint, submissionID: bigint): Promise<IGrpcResponse<Submission>> {
        const request = new SubmissionReviewersRequest({
            courseID: courseID,
            submissionID: submissionID
        })
        return this.grpcSend<Submission>(this.agService.getSubmission, request)
    }

    public getGroupSubmissions(courseID: bigint, groupID: bigint): Promise<IGrpcResponse<Submissions>> {
        const request = new SubmissionRequest({
            courseID: courseID,
            groupID: groupID,
        })
        return this.grpcSend<Submissions>(this.agService.getSubmissions, request)
    }

    public getSubmissionsByCourse(courseID: bigint, type: SubmissionsForCourseRequest_Type): Promise<IGrpcResponse<CourseSubmissions>> {
        const request = new SubmissionsForCourseRequest({
            courseID: courseID,
            type: type,
        })
        return this.grpcSend<CourseSubmissions>(this.agService.getSubmissionsByCourse, request)
    }

    public updateSubmission(courseID: bigint, s: Submission): Promise<IGrpcResponse<Void>> {
        const request = new UpdateSubmissionRequest({
            courseID: courseID,
            submissionID: s.ID,
            status: s.status,
            released: s.released,
            score: s.score,
        })
        return this.grpcSend<Void>(this.agService.updateSubmission, request)
    }

    public updateSubmissions(assignmentID: bigint, courseID: bigint, score: number, release: boolean, approve: boolean): Promise<IGrpcResponse<Void>> {
        const request = new UpdateSubmissionsRequest({
            courseID: courseID,
            assignmentID: assignmentID,
            scoreLimit: score,
            release: release,
            approve: approve,
        })
        return this.grpcSend<Void>(this.agService.updateSubmissions, request)
    }

    public rebuildSubmission(assignmentID: bigint, submissionID: bigint, courseID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new RebuildRequest({
            courseID: courseID,
            assignmentID: assignmentID,
            submissionID: submissionID,
        })
        return this.grpcSend<Void>(this.agService.rebuildSubmissions, request)
    }

    public rebuildSubmissions(assignmentID: bigint, courseID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new RebuildRequest({
            courseID: courseID,
            assignmentID: assignmentID,
        })
        return this.grpcSend<Void>(this.agService.rebuildSubmissions, request)
    }

    // /* MANUAL GRADING */ //

    public createBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<GradingBenchmark>> {
        return this.grpcSend<GradingBenchmark>(this.agService.createBenchmark, bm)
    }

    public createCriterion(c: GradingCriterion): Promise<IGrpcResponse<GradingCriterion>> {
        return this.grpcSend<GradingCriterion>(this.agService.createCriterion, c)
    }

    public updateBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateBenchmark, bm)
    }

    public updateCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateCriterion, c)
    }

    public deleteBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.deleteBenchmark, bm)
    }

    public deleteCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.deleteCriterion, c)
    }

    public createReview(r: Review, courseID: bigint): Promise<IGrpcResponse<Review>> {
        const request = new ReviewRequest({
            courseID: courseID,
            review: r,
        })
        return this.grpcSend<Review>(this.agService.createReview, request)
    }

    public updateReview(r: Review, courseID: bigint): Promise<IGrpcResponse<Review>> {
        const request = new ReviewRequest({
            courseID: courseID,
            review: r,
        })
        return this.grpcSend<Review>(this.agService.updateReview, request)
    }

    public getReviewers(submissionID: bigint, courseID: bigint): Promise<IGrpcResponse<Reviewers>> {
        const request = new SubmissionReviewersRequest({
            courseID: courseID,
            submissionID: submissionID,
        })
        return this.grpcSend<Reviewers>(this.agService.getReviewers, request)
    }

    // /* REPOSITORY */ //

    public getRepositories(courseID: bigint, types: Repository_Type[]): Promise<IGrpcResponse<Repositories>> {
        const req = new URLRequest({
            courseID: courseID,
            repoTypes: types,
        })
        return this.grpcSend<Repositories>(this.agService.getRepositories, req)
    }

    // /* ORGANIZATIONS */ //

    public getOrganization(orgName: string): Promise<IGrpcResponse<Organization>> {
        const request = new OrgRequest({
            orgName: orgName,
        })
        return this.grpcSend<Organization>(this.agService.getOrganization, request)
    }

    public isEmptyRepo(courseID: bigint, userID: bigint, groupID: bigint): Promise<IGrpcResponse<Void>> {
        const request = new RepositoryRequest({
            courseID: courseID,
            userID: userID,
            groupID: groupID,
        })
        return this.grpcSend<Void>(this.agService.isEmptyRepo, request)
    }

    private grpcSend<T>(method: any, request: any): Promise<IGrpcResponse<T>> {
        const grpcPromise = new Promise<IGrpcResponse<T>>((resolve) => {
            method.call(this.agService, request,
                (err: ConnectError, response: T) => {
                    if (err) {
                        if (err.code > 0) {
                            const code = new Status({
                                Code: BigInt(err.code),
                                Error: err.message,
                            })
                            const temp: IGrpcResponse<T> = {
                                status: code,
                            }
                            this.logErr(temp, method.name)
                            resolve(temp)
                        }
                    } else {
                        const code = new Status({
                            Code: BigInt(0),
                            Error: "",
                        })
                        const temp: IGrpcResponse<T> = {
                            data: response as T,
                            status: code,
                        }
                        resolve(temp)
                    }
                })
        })
        return grpcPromise
    }

    // logErr logs any gRPC error to the console.
    private logErr(resp: IGrpcResponse<any>, methodName: string): void {
        if (resp.status.Code !== BigInt(0)) {
            console.log("GRPC " + methodName + " failed with code "
                + resp.status.Code + ": " + resp.status.Error)
        }
    }
}
