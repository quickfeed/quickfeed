import * as grpcWeb from "grpc-web";
import {
    Assignments,
    AuthorizationResponse,
    Benchmarks,
    Course,
    CourseRequest,
    CourseSubmissions,
    Courses,
    EnrollmentStatusRequest,
    Enrollment,
    EnrollmentRequest,
    Enrollments,
    GetGroupRequest,
    GradingBenchmark,
    GradingCriterion,
    Group,
    GroupRequest,
    Groups,
    Organization,
    OrgRequest,
    Providers,
    RebuildRequest,
    Repositories,
    Repository,
    RepositoryRequest,
    Review,
    ReviewRequest,
    Status,
    SubmissionRequest,
    SubmissionsForCourseRequest,
    Submission,
    Submissions,
    SubmissionReviewersRequest,
    UpdateSubmissionRequest,
    UpdateSubmissionsRequest,
    URLRequest,
    User,
    Users,
    Void,
    Reviewers,
} from "../../proto/ag_pb";
import { AutograderServiceClient } from "../../proto/AgServiceClientPb";
import { UserManager } from "./UserManager";
import { ISubmission } from "../models";
import { LoadCriteriaRequest } from '../../proto/ag_pb';

export interface IGrpcResponse<T> {
    status: Status;
    data?: T;
}

export class GrpcManager {

    private agService: AutograderServiceClient;
    private userMan: UserManager;

    constructor() {
        // to test on localhost via port forwarding, use make local Makefile target
        this.agService = new AutograderServiceClient("https://" + window.location.hostname, null, null);
    }

    public setUserMan(man: UserManager) {
        this.userMan = man;
    }

    // /* USERS */ //

    public getUser(): Promise<IGrpcResponse<User>> {
        return this.grpcSend<User>(this.agService.getUser, new Void());
    }

    public getUsers(): Promise<IGrpcResponse<Users>> {
        return this.grpcSend<Users>(this.agService.getUsers, new Void());
    }

    public updateUser(user: User): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateUser, user);
    }

    public isAuthorizedTeacher(): Promise<IGrpcResponse<AuthorizationResponse>> {
        return this.grpcSend<AuthorizationResponse>(this.agService.isAuthorizedTeacher, new Void());
    }

    // /* COURSES */ //

    public createCourse(course: Course): Promise<IGrpcResponse<Course>> {
        return this.grpcSend<Course>(this.agService.createCourse, course);
    }

    public updateCourse(course: Course): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateCourse, course);
    }

    public getCourse(courseID: number): Promise<IGrpcResponse<Course>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Course>(this.agService.getCourse, request);
    }

    public getCourses(): Promise<IGrpcResponse<Courses>> {
        return this.grpcSend<Courses>(this.agService.getCourses, new Void());
    }

    public getCoursesByUser(userID: number, statuses: Enrollment.UserStatus[]): Promise<IGrpcResponse<Courses>> {
        const request = new EnrollmentStatusRequest();
        request.setUserid(userID);
        request.setStatusesList(statuses);
        return this.grpcSend<Courses>(this.agService.getCoursesByUser, request);
    }

    public updateCourseVisibility(request: Enrollment): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateCourseVisibility, request)
    }

    // /* ASSIGNMENTS */ //

    public getAssignments(courseID: number): Promise<IGrpcResponse<Assignments>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Assignments>(this.agService.getAssignments, request);
    }

    public updateAssignments(courseID: number): Promise<IGrpcResponse<Void>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.updateAssignments, request);
    }

    // /* ENROLLMENTS */ //

    public getEnrollmentsByUser(userID: number, statuses?: Enrollment.UserStatus[]): Promise<IGrpcResponse<Enrollments>> {
        const request = new EnrollmentStatusRequest();
        request.setUserid(userID);
        request.setStatusesList(statuses ?? []);
        return this.grpcSend<Enrollments>(this.agService.getEnrollmentsByUser, request);
    }

    public getEnrollmentsByCourse(courseID: number, withoutGroupMembers?: boolean, withActivity?: boolean, statuses?: Enrollment.UserStatus[]):
        Promise<IGrpcResponse<Enrollments>> {
        const request = new EnrollmentRequest();
        request.setCourseid(courseID);
        request.setIgnoregroupmembers(withoutGroupMembers ?? false);
        request.setWithactivity(withActivity ?? false);
        request.setStatusesList(statuses ?? []);
        return this.grpcSend<Enrollments>(this.agService.getEnrollmentsByCourse, request);
    }

    public createEnrollment(courseID: number, userID: number): Promise<IGrpcResponse<Void>> {
        const request = new Enrollment();
        request.setUserid(userID);
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.createEnrollment, request);
    }

    public updateEnrollment(request: Enrollment): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateEnrollment, request);
    }

    public updateEnrollments(courseID: number): Promise<IGrpcResponse<Void>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.updateEnrollments, request);
    }

    // /* GROUPS */ //

    public getGroup(groupID: number): Promise<IGrpcResponse<Group>> {
        const request = new GetGroupRequest();
        request.setGroupid(groupID);
        return this.grpcSend<Group>(this.agService.getGroup, request);
    }

    public getGroupByUserAndCourse(courseID: number, userID: number): Promise<IGrpcResponse<Group>> {
        const request = new GroupRequest();
        request.setUserid(userID);
        request.setCourseid(courseID);
        return this.grpcSend<Group>(this.agService.getGroupByUserAndCourse, request);
    }

    public getGroupsByCourse(courseID: number): Promise<IGrpcResponse<Groups>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Groups>(this.agService.getGroupsByCourse, request);
    }

    public updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<IGrpcResponse<Void>> {
        const request = new Group();
        request.setId(groupID);
        request.setStatus(status);
        return this.grpcSend<Void>(this.agService.updateGroup, request);
    }

    public updateGroup(group: Group): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateGroup, group);
    }

    public deleteGroup(courseID: number, groupID: number): Promise<IGrpcResponse<Void>> {
        const request = new GroupRequest();
        request.setGroupid(groupID);
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.deleteGroup, request);
    }

    public createGroup(courseID: number, name: string, users: number[]): Promise<IGrpcResponse<Group>> {
        const request = new Group();
        request.setName(name);
        request.setCourseid(courseID);
        const groupUsers: User[] = [];
        users.forEach((ele) => {
            const usr = new User();
            usr.setId(ele);
            groupUsers.push(usr);
        });
        request.setUsersList(groupUsers);
        return this.grpcSend<Group>(this.agService.createGroup, request);
    }

    // /* SUBMISSIONS */ //

    public getSubmissions(courseID: number, userID: number): Promise<IGrpcResponse<Submissions>> {
        const request = new SubmissionRequest();
        request.setCourseid(courseID);
        request.setUserid(userID);
        return this.grpcSend<Submissions>(this.agService.getSubmissions, request);
    }

    public getGroupSubmissions(courseID: number, groupID: number): Promise<IGrpcResponse<Submissions>> {
        const request = new SubmissionRequest();
        request.setCourseid(courseID);
        request.setGroupid(groupID);
        return this.grpcSend<Submissions>(this.agService.getSubmissions, request);
    }

    public getSubmissionsByCourse(courseID: number, type: SubmissionsForCourseRequest.Type): Promise<IGrpcResponse<CourseSubmissions>> {
        const request = new SubmissionsForCourseRequest();
        request.setCourseid(courseID);
        request.setType(type);
        return this.grpcSend<CourseSubmissions>(this.agService.getSubmissionsByCourse, request);
    }

    public rebuildSubmission(assignmentID: number, submissionID: number): Promise<IGrpcResponse<Submission>> {
        const request = new RebuildRequest();
        request.setAssignmentid(assignmentID);
        request.setSubmissionid(submissionID);
        return this.grpcSend<Submission>(this.agService.rebuildSubmission, request);
    }

    public updateSubmission(courseID: number, s: ISubmission): Promise<IGrpcResponse<Void>> {
        const request = new UpdateSubmissionRequest();
        request.setSubmissionid(s.id);
        request.setCourseid(courseID);
        request.setStatus(s.status);
        request.setReleased(s.released);
        request.setScore(s.score);
        return this.grpcSend<Void>(this.agService.updateSubmission, request);
    }

    public updatesubmissions(assignmentID: number, courseID: number, score: number, release: boolean, approve: boolean) {
        const request = new UpdateSubmissionsRequest();
        request.setAssignmentid(assignmentID);
        request.setCourseid(courseID);
        request.setScorelimit(score);
        request.setRelease(release);
        request.setApprove(approve);
        return this.grpcSend<Void>(this.agService.updateSubmissions, request);
    }

    // /* MANUAL GRADING */ //

    public createBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<GradingBenchmark>> {
        return this.grpcSend<GradingBenchmark>(this.agService.createBenchmark, bm);
    }

    public createCriterion(c: GradingCriterion): Promise<IGrpcResponse<GradingCriterion>> {
        return this.grpcSend<GradingCriterion>(this.agService.createCriterion, c);
    }

    public updateBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateBenchmark, bm);
    }

    public updateCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateCriterion, c);
    }

    public deleteBenchmark(bm: GradingBenchmark): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.deleteBenchmark, bm);
    }

    public deleteCriterion(c: GradingCriterion): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.deleteCriterion, c);
    }

    public createReview(r: Review, courseID: number): Promise<IGrpcResponse<Review>> {
        const request = new ReviewRequest();
        request.setReview(r);
        request.setCourseid(courseID);
        return this.grpcSend<Review>(this.agService.createReview, request);
    }

    public updateReview(r: Review, courseID: number): Promise<IGrpcResponse<Void>> {
        const request = new ReviewRequest();
        request.setReview(r);
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.updateReview, request);
    }

    public getReviewers(submissionID: number, courseID: number): Promise<IGrpcResponse<Reviewers>> {
        const request = new SubmissionReviewersRequest();
        request.setSubmissionid(submissionID);
        request.setCourseid(courseID);
        return this.grpcSend<Reviewers>(this.agService.getReviewers, request);
    }

    public loadCriteria(assignmentID: number, courseID: number): Promise<IGrpcResponse<Benchmarks>> {
        const request = new LoadCriteriaRequest();
        request.setAssignmentid(assignmentID);
        request.setCourseid(courseID);
        return this.grpcSend<Benchmarks>(this.agService.loadCriteria, request);
    }

    // /* REPOSITORY */ //

    public getRepositories(courseID: number, types: Repository.Type[]): Promise<IGrpcResponse<Repositories>> {
        const req = new URLRequest();
        req.setCourseid(courseID);
        req.setRepotypesList(types);
        return this.grpcSend<Repositories>(this.agService.getRepositories, req);
    }

    // /* ORGANIZATIONS */ //

    public getOrganization(orgName: string): Promise<IGrpcResponse<Organization>> {
        const request = new OrgRequest();
        request.setOrgname(orgName);
        return this.grpcSend<Organization>(this.agService.getOrganization, request);
    }

    public getProviders(): Promise<IGrpcResponse<Providers>> {
        return this.grpcSend<Providers>(this.agService.getProviders, new Void());
    }

    public isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<IGrpcResponse<Void>> {
        const request = new RepositoryRequest();
        request.setUserid(userID);
        request.setGroupid(groupID);
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.isEmptyRepo, request);
    }

    // /* UTILITY */ //

    private grpcSend<T>(method: any, request: any): Promise<IGrpcResponse<T>> {
        const grpcPromise = new Promise<IGrpcResponse<T>>((resolve) => {
            let userID = "";
            // currentUser reference is created on authorization with a provider and stores a User object.
            // This object can be used for user validation. This implementation sends user ID to simplify
            // and standardize different server checks.
            const currentUser = this.userMan.getCurrentUser();
            if (currentUser != null) {
                userID = currentUser.getId().toString();
            }
            method.call(this.agService, request, { "custom-header-1": "value1", "user": userID },
                (err: grpcWeb.Error, response: T) => {
                    if (err) {
                        if (err.code !== grpcWeb.StatusCode.OK) {
                            const code = new Status();
                            code.setCode(err.code);
                            code.setError(err.message);
                            const temp: IGrpcResponse<T> = {
                                status: code,
                            };
                            this.logErr(temp, method.name);
                            resolve(temp);
                        }
                    } else {
                        const code = new Status();
                        code.setCode(0);
                        code.setError("OK");
                        const temp: IGrpcResponse<T> = {
                            data: response as T,
                            status: code,
                        };
                        resolve(temp);
                    }
                });
        });
        return grpcPromise;
    }

    // logErr logs any gRPC error to the console.
    private logErr(resp: IGrpcResponse<any>, methodName: string): void {
        if (resp.status.getCode() !== 0) {
            console.log("GRPC " + methodName + " failed with code "
                + resp.status.getCode() + ": " + resp.status.getError());
        }
    }
}
