import * as grpcWeb from "grpc-web";
import {
    Assignments,
    AuthorizationResponse,
    Course,
    CourseRequest,
    Courses,
    CoursesListRequest,
    Enrollment,
    EnrollmentRequest,
    Enrollments,
    GetGroupRequest,
    Group,
    GroupRequest,
    Groups,
    LabRequest,
    LabResultLinks,
    Organization,
    OrgRequest,
    Providers,
    Repositories,
    Repository,
    RepositoryRequest,
    Status,
    SubmissionRequest,
    Submissions,
    UpdateSubmissionRequest,
    URLRequest,
    User,
    Users,
    Void,
} from "../../proto/ag_pb";
import { AutograderServiceClient } from "../../proto/AgServiceClientPb";
import { UserManager } from "./UserManager";

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
        const request = new Void();
        return this.grpcSend<User>(this.agService.getUser, request);
    }

    public getUsers(): Promise<IGrpcResponse<Users>> {
        const request = new Void();
        return this.grpcSend<Users>(this.agService.getUsers, request);
    }

    public updateUser(user: User, isAdmin?: boolean): Promise<IGrpcResponse<User>> {
        const requrest = new User();
        requrest.setId(user.getId());
        requrest.setAvatarurl(user.getAvatarurl());
        requrest.setEmail(user.getEmail());
        requrest.setName(user.getName());
        requrest.setStudentid(user.getStudentid());
        if (isAdmin) {
            requrest.setIsadmin(isAdmin);
        } else {
            requrest.setIsadmin(user.getIsadmin());
        }
        return this.grpcSend(this.agService.updateUser, requrest);
    }

    public isAuthorizedTeacher(): Promise<IGrpcResponse<AuthorizationResponse>> {
        const voidy = new Void();
        return this.grpcSend<AuthorizationResponse>(this.agService.isAuthorizedTeacher, voidy);
    }

    // /* COURSES */ //

    public createCourse(course: Course): Promise<IGrpcResponse<Course>> {
        return this.grpcSend<Course>(this.agService.createCourse, course);
    }

    public updateCourse(course: Course): Promise<IGrpcResponse<Course>> {
        return this.grpcSend<Course>(this.agService.updateCourse, course);
    }

    public getCourse(courseID: number): Promise<IGrpcResponse<Course>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Course>(this.agService.getCourse, request);
    }

    public getCourses(): Promise<IGrpcResponse<Courses>> {
        const request = new Void();
        return this.grpcSend<Courses>(this.agService.getCourses, request);
    }

    public getCoursesWithEnrollment(userID: number, state: Enrollment.UserStatus[]): Promise<IGrpcResponse<Courses>> {
        const request = new CoursesListRequest();
        request.setUserid(userID);
        request.setStatesList(state);
        return this.grpcSend<Courses>(this.agService.getCoursesWithEnrollment, request);
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

    public getEnrollmentsByCourse(courseID: number, noGroupMembers?: boolean, state?: any):
        Promise<IGrpcResponse<Enrollments>> {

        const request = new EnrollmentRequest();
        request.setCourseid(courseID);
        if (noGroupMembers) {
            request.setFilteroutgroupmembers(noGroupMembers);
        }
        if (state) {
            request.setStatesList(state);
        }
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

    public getGroups(courseID: number): Promise<IGrpcResponse<Groups>> {
        const request = new CourseRequest();
        request.setCourseid(courseID);
        return this.grpcSend<Groups>(this.agService.getGroups, request);
    }

    public updateGroupStatus(groupID: number, state: Group.GroupStatus): Promise<IGrpcResponse<Void>> {
        const request = new Group();
        request.setId(groupID);
        request.setStatus(state);
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

    public getCourseLabSubmissions(courseID: number, groupLabs: boolean): Promise<IGrpcResponse<LabResultLinks>> {
        const request = new LabRequest();
        request.setCourseid(courseID);
        request.setGrouplabs(groupLabs);
        return this.grpcSend<LabResultLinks>(this.agService.getCourseLabSubmissions, request);
    }

    public rebuildSubmission(assignmentID: number, submissionID: number): Promise<IGrpcResponse<Void>> {
        const request = new LabRequest();
        request.setAssignmentid(assignmentID);
        request.setSubmissionid(submissionID);
        return this.grpcSend<Void>(this.agService.rebuildSubmission, request);
    }

    public updateSubmission(courseID: number, submissionID: number, approve: boolean): Promise<IGrpcResponse<Void>> {
        const request = new UpdateSubmissionRequest();
        request.setSubmissionid(submissionID);
        request.setCourseid(courseID);
        request.setApprove(approve);
        return this.grpcSend<Void>(this.agService.updateSubmission, request);
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
        const request = new Void();
        return this.grpcSend<Providers>(this.agService.getProviders, request);
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
                (err: grpcWeb.Error, response: T | undefined) => {
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
