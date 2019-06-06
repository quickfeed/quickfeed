
import * as grpcWeb from "grpc-web";

import { AutograderServiceClient } from "../../proto/AgServiceClientPb";
import {
    RepositoryRequest,
    Providers,
    User,
    Users,
    RecordRequest,
    Course,
    Courses,
    Submission,
    Submissions,
    Assignments,
    ActionRequest,
    Enrollments,
    Group,
    Groups,
    Void,
    Directories,
    DirectoryRequest,
    URLResponse,
} from "../../proto/ag_pb";
import { INewGroup } from "../models";
import { UserManager } from "./UserManager";

export interface IGrpcResponse<T> {
    statusCode: number;
    data?: T;
}

export class GrpcManager {

    agService: AutograderServiceClient;
    private userMan: UserManager;

    constructor() {
        this.agService = new AutograderServiceClient("http://localhost:8080", null, null);
    }

    public setUserMan(man: UserManager) {
        this.userMan = man;
    }

    // /* USERS */ //
/*
    public getSelf(): Promise<IGrpcResponse<User>> {
        const request = new Void();
        return this.grpcSend<User>(this.agService.getSelf, request);
    }*/

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

    // /* COURSES */ //

    public createCourse(course: Course): Promise<IGrpcResponse<Course>> {
        console.log("GRPCMan calls createCourse for course " + course.getName());
        return this.grpcSend<Course>(this.agService.createCourse, course);
    }

    public updateCourse(course: Course): Promise<IGrpcResponse<Course>> {
        console.log("GRPCMan calls updateCourse for course " + course.getName());
        return this.grpcSend<Course>(this.agService.updateCourse, course);
    }

    public refreshCourse(courseID: number): Promise<IGrpcResponse<any>> {
        const request = new RecordRequest();
        request.setId(courseID);
        return this.grpcSend<Assignments>(this.agService.refreshCourse, request);
    }

    public getCourse(id: number): Promise<IGrpcResponse<Course>> {
        const request = new RecordRequest();
        request.setId(id);
        return this.grpcSend(this.agService.getCourse, request);
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

    public getCourseInformationURL(courseID: number): Promise<IGrpcResponse<URLResponse>> {
        const request = new RecordRequest();
        request.setId(courseID);
        return this.grpcSend<URLResponse>(this.agService.getCourseInformationURL, request);
    }

    // /* ASSIGNMENTS */ //

    public getAssignments(courseId: number): Promise<IGrpcResponse<Assignments>> {
        const request = new RecordRequest();
        request.setId(courseId);
        return this.grpcSend<Assignments>(this.agService.getAssignments, request);
    }

    // /* ENROLLMENTS */ //

    public getEnrollmentsByCourse(courseid: number, state: any): Promise<IGrpcResponse<Enrollments>> {
        const request = new RecordRequest();
        request.setId(courseid);
        request.setStatusesList(state);
        return this.grpcSend<Enrollments>(this.agService.getEnrollmentsByCourse, request);
    }

    public createEnrollment(userid: number, courseid: number): Promise<IGrpcResponse<Void>> {
        const request = new ActionRequest();
        request.setUserid(userid);
        request.setCourseid(courseid);
        return this.grpcSend<Void>(this.agService.createEnrollment, request);
    }

    public updateEnrollment(userid: number, courseid: number, state: any): Promise<IGrpcResponse<Void>> {
        const request = new ActionRequest();
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
        const request = new ActionRequest();
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
        const request = new Group();
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
        const request = new ActionRequest();
        request.setCourseid(courseID);
        request.setUserid(userID);
        return this.grpcSend<Submissions>(this.agService.getSubmissions, request);
    }

    public getGroupSubmissions(courseID: number, groupID: number): Promise<IGrpcResponse<Submissions>> {
        const request = new ActionRequest();
        request.setCourseid(courseID);
        request.setGroupid(groupID);
        return this.grpcSend<Submissions>(this.agService.getGroupSubmissions, request);
    }

    public updateSubmission(submissionID: number): Promise<IGrpcResponse<Void>> {
        const request = new RecordRequest();
        request.setId(submissionID);
        return this.grpcSend<Void>(this.agService.updateSubmission, request);
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

    public getDirectories(provider: string): Promise<IGrpcResponse<Directories>> {
        const request = new DirectoryRequest();
        request.setProvider(provider);
        return this.grpcSend<Directories>(this.agService.getDirectories, request);
    }

    // /* UTILITY */ //

    private grpcSend<T>(method: any, request: any): Promise<IGrpcResponse<T>> {

        const grpcPromise = new Promise<IGrpcResponse<T>>((resolve) => {

            let userId = "";

            // currentUser reference is created on authorization with a provider and stores a User object
            // This object can be used for user validation. This implementation sends user ID to simplify
            // and standardize different server checks.
            // Alternative solution is to send the token, which requires a sequre way of storing the token
            const currentUser = this.userMan.getCurrentUser();
            if (currentUser != null) {
                userId = currentUser.getId().toString();
            }

            const call = method.call(this.agService, request, { "custom-header-1": "value1", "user": userId },
                (err: grpcWeb.Error, response: T | undefined) => {
                    if (err) {
                        if (err.code !== grpcWeb.StatusCode.OK) {
                            const temp: IGrpcResponse<T> = {
                                statusCode: err.code,
                            };
                            resolve(temp);
                        }
                    } else {
                        const temp: IGrpcResponse<T> = {
                            data: response as T,
                            statusCode: 0,
                        };
                        resolve(temp);
                    }
                });
        });
        return grpcPromise;
    }
}
