import * as grpcWeb from "grpc-web";
import {
    ApproveSubmissionRequest,
    Assignments,
    AuthorizationResponse,
    Course,
    Courses,
    Enrollment,
    EnrollmentRequest,
    Enrollments,
    Group,
    GroupRequest,
    Groups,
    Organizations,
    Provider,
    Providers,
    RecordRequest,
    RepositoryRequest,
    Status,
    Submission,
    SubmissionRequest,
    Submissions,
    URLResponse,
    User,
    Users,
    Void,
} from "../../proto/ag_pb";
import { AutograderServiceClient } from "../../proto/AgServiceClientPb";
import { INewGroup } from "../models";
import { UserManager } from "./UserManager";

// will have either data or message
export interface IGrpcResponse<T> {
    status: Status;
    data?: T;
}

export class GrpcManager {

    private agService: AutograderServiceClient;
    private userMan: UserManager;

    constructor() {
        this.agService = new AutograderServiceClient("http://localhost:8080", null, null);
    }

    public setUserMan(man: UserManager) {
        this.userMan = man;
    }

    // /* USERS */ //

    public getUsers(): Promise<IGrpcResponse<Users>> {
        const request = new Void();
        return this.grpcSend<Users>(this.agService.getUsers, request);
    }

    public getUser(id: number): Promise<IGrpcResponse<User>> {
        const request = new RecordRequest();
        request.setId(id);
        return this.grpcSend<User>(this.agService.getUser, request);
    }

    public updateUser(user: User, isadmin?: boolean): Promise<IGrpcResponse<User>> {
        const requrest = new User();
        requrest.setId(user.getId());
        requrest.setAvatarurl(user.getAvatarurl());
        requrest.setEmail(user.getEmail());
        requrest.setName(user.getName());
        requrest.setStudentid(user.getStudentid());
        if (isadmin) {
            requrest.setIsadmin(isadmin);
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

    public updateAssignments(courseID: number): Promise<IGrpcResponse<Void>> {
        const request = new RecordRequest();
        request.setId(courseID);
        return this.grpcSend<Void>(this.agService.updateAssignments, request);
    }

    public getCourse(id: number): Promise<IGrpcResponse<Course>> {
        const request = new RecordRequest();
        request.setId(id);
        return this.grpcSend<Course>(this.agService.getCourse, request);
    }

    public getCourses(): Promise<IGrpcResponse<Courses>> {
        const request = new Void();
        return this.grpcSend<Courses>(this.agService.getCourses, request);
    }

    public getCoursesWithEnrollment(userid: number, state: any): Promise<IGrpcResponse<Courses>> {
        const request = new RecordRequest();
        request.setId(userid);
        request.setStatusesList(state);
        return this.grpcSend<Courses>(this.agService.getCoursesWithEnrollment, request);
    }

    // /* ASSIGNMENTS */ //

    public getAssignments(courseId: number): Promise<IGrpcResponse<Assignments>> {
        const request = new RecordRequest();
        request.setId(courseId);
        return this.grpcSend<Assignments>(this.agService.getAssignments, request);
    }

    // /* ENROLLMENTS */ //

    public getEnrollmentsByCourse(courseid: number, noGroupMembers?: boolean, state?: any):
        Promise<IGrpcResponse<Enrollments>> {

        const request = new EnrollmentRequest();
        request.setCourseid(courseid);
        if (noGroupMembers) {
            request.setFilteroutgroupmembers(noGroupMembers);
        }
        if (state) {
            request.setStatesList(state);
        }
        return this.grpcSend<Enrollments>(this.agService.getEnrollmentsByCourse, request);
    }

    public createEnrollment(userid: number, courseid: number): Promise<IGrpcResponse<Void>> {
        const request = new Enrollment();
        request.setUserid(userid);
        request.setCourseid(courseid);
        return this.grpcSend<Void>(this.agService.createEnrollment, request);
    }

    public updateEnrollment(userid: number, courseid: number, state: any): Promise<IGrpcResponse<Void>> {
        const request = new Enrollment();
        request.setUserid(userid);
        request.setCourseid(courseid);
        request.setStatus(state);
        return this.grpcSend<Void>(this.agService.updateEnrollment, request);
    }

    // /* GROUPS */ //

    public getGroup(groupID: number): Promise<IGrpcResponse<Group>> {
        const request = new RecordRequest();
        request.setId(groupID);
        return this.grpcSend<Group>(this.agService.getGroup, request);
    }

    public getGroupByUserAndCourse(userID: number, courseID: number): Promise<IGrpcResponse<Group>> {
        const request = new GroupRequest();
        request.setUserid(userID);
        request.setCourseid(courseID);
        return this.grpcSend<Group>(this.agService.getGroupByUserAndCourse, request);
    }

    public getGroups(courseid: number): Promise<IGrpcResponse<Groups>> {
        const request = new RecordRequest();
        request.setId(courseid);
        return this.grpcSend<Groups>(this.agService.getGroups, request);
    }

    public updateGroupStatus(groupid: number, state: Group.GroupStatus): Promise<IGrpcResponse<Void>> {
        const request = new Group();
        request.setId(groupid);
        request.setStatus(state);
        return this.grpcSend<Void>(this.agService.updateGroup, request);
    }

    public updateGroup(grp: Group): Promise<IGrpcResponse<Void>> {
        return this.grpcSend<Void>(this.agService.updateGroup, grp);
    }

    public deleteGroup(groupid: number): Promise<IGrpcResponse<Void>> {
        const request = new RecordRequest();
        request.setId(groupid);
        return this.grpcSend<Void>(this.agService.deleteGroup, request);
    }

    public createGroup(igrp: INewGroup, courseid: number): Promise<IGrpcResponse<Group>> {
        const request = new Group();
        request.setName(igrp.name);
        request.setCourseid(courseid);
        const grpusers: User[] = [];
        igrp.userids.forEach((ele) => {
            const usr = new User();
            usr.setId(ele);
            grpusers.push(usr);
        });
        request.setUsersList(grpusers);
        return this.grpcSend<Group>(this.agService.createGroup, request);
    }

    // /* SUBMISSIONS */ //

    public getSubmission(assignmentID: number): Promise<IGrpcResponse<Submission>> {
        const request = new RecordRequest();
        request.setId(assignmentID);
        return this.grpcSend<Submission>(this.agService.getSubmission, request);
    }

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

    public approveSubmission(submissionID: number, courseID: number): Promise<IGrpcResponse<Void>> {
        const request = new ApproveSubmissionRequest();
        request.setSubmissionid(submissionID);
        request.setCourseid(courseID);
        return this.grpcSend<Void>(this.agService.approveSubmission, request);
    }

    // /* REPOSITORY */ //

    public getRepositoryURL(courseid: number, repotype: number): Promise<IGrpcResponse<URLResponse>> {
        const request = new RepositoryRequest();
        request.setCourseid(courseid);
        request.setType(repotype);
        return this.grpcSend<URLResponse>(this.agService.getRepositoryURL, request);
    }

    public getProviders(): Promise<IGrpcResponse<Providers>> {
        const request = new Void();
        return this.grpcSend<Providers>(this.agService.getProviders, request);
    }

    public getOrganizations(provider: string): Promise<IGrpcResponse<Organizations>> {
        const request = new Provider();
        request.setProvider(provider);
        return this.grpcSend<Organizations>(this.agService.getOrganizations, request);
    }

    // /* UTILITY */ //

    private grpcSend<T>(method: any, request: any): Promise<IGrpcResponse<T>> {
        const grpcPromise = new Promise<IGrpcResponse<T>>((resolve) => {
            let userId = "";
            // currentUser reference is created on authorization with a provider and stores a User object.
            // This object can be used for user validation. This implementation sends user ID to simplify
            // and standardize different server checks.
            // Alternative solution is to send the token, which requires a secure way of storing the token.
            const currentUser = this.userMan.getCurrentUser();
            if (currentUser != null) {
                userId = currentUser.getId().toString();
            }

            method.call(this.agService, request, { "custom-header-1": "value1", "user": userId },
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
                        // TODO(vera): this can be handled in another way, needs synchronization with backend
                        code.setError("");
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
