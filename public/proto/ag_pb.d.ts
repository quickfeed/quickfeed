import * as jspb from "google-protobuf"


export class User extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getIsadmin(): boolean;
  setIsadmin(value: boolean): void;

  getName(): string;
  setName(value: string): void;

  getStudentid(): string;
  setStudentid(value: string): void;

  getEmail(): string;
  setEmail(value: string): void;

  getAvatarurl(): string;
  setAvatarurl(value: string): void;

  getLogin(): string;
  setLogin(value: string): void;

  getRemoteidentitiesList(): Array<RemoteIdentity>;
  setRemoteidentitiesList(value: Array<RemoteIdentity>): void;
  clearRemoteidentitiesList(): void;
  addRemoteidentities(value?: RemoteIdentity, index?: number): RemoteIdentity;

  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): void;
  clearEnrollmentsList(): void;
  addEnrollments(value?: Enrollment, index?: number): Enrollment;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): User.AsObject;
  static toObject(includeInstance: boolean, msg: User): User.AsObject;
  static serializeBinaryToWriter(message: User, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): User;
  static deserializeBinaryFromReader(message: User, reader: jspb.BinaryReader): User;
}

export namespace User {
  export type AsObject = {
    id: number,
    isadmin: boolean,
    name: string,
    studentid: string,
    email: string,
    avatarurl: string,
    login: string,
    remoteidentitiesList: Array<RemoteIdentity.AsObject>,
    enrollmentsList: Array<Enrollment.AsObject>,
  }
}

export class Users extends jspb.Message {
  getUsersList(): Array<User>;
  setUsersList(value: Array<User>): void;
  clearUsersList(): void;
  addUsers(value?: User, index?: number): User;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Users.AsObject;
  static toObject(includeInstance: boolean, msg: Users): Users.AsObject;
  static serializeBinaryToWriter(message: Users, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Users;
  static deserializeBinaryFromReader(message: Users, reader: jspb.BinaryReader): Users;
}

export namespace Users {
  export type AsObject = {
    usersList: Array<User.AsObject>,
  }
}

export class RemoteIdentity extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getProvider(): string;
  setProvider(value: string): void;

  getRemoteid(): number;
  setRemoteid(value: number): void;

  getAccesstoken(): string;
  setAccesstoken(value: string): void;

  getUserid(): number;
  setUserid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RemoteIdentity.AsObject;
  static toObject(includeInstance: boolean, msg: RemoteIdentity): RemoteIdentity.AsObject;
  static serializeBinaryToWriter(message: RemoteIdentity, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RemoteIdentity;
  static deserializeBinaryFromReader(message: RemoteIdentity, reader: jspb.BinaryReader): RemoteIdentity;
}

export namespace RemoteIdentity {
  export type AsObject = {
    id: number,
    provider: string,
    remoteid: number,
    accesstoken: string,
    userid: number,
  }
}

export class Group extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getName(): string;
  setName(value: string): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  getTeamid(): number;
  setTeamid(value: number): void;

  getStatus(): Group.GroupStatus;
  setStatus(value: Group.GroupStatus): void;

  getUsersList(): Array<User>;
  setUsersList(value: Array<User>): void;
  clearUsersList(): void;
  addUsers(value?: User, index?: number): User;

  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): void;
  clearEnrollmentsList(): void;
  addEnrollments(value?: Enrollment, index?: number): Enrollment;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Group.AsObject;
  static toObject(includeInstance: boolean, msg: Group): Group.AsObject;
  static serializeBinaryToWriter(message: Group, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Group;
  static deserializeBinaryFromReader(message: Group, reader: jspb.BinaryReader): Group;
}

export namespace Group {
  export type AsObject = {
    id: number,
    name: string,
    courseid: number,
    teamid: number,
    status: Group.GroupStatus,
    usersList: Array<User.AsObject>,
    enrollmentsList: Array<Enrollment.AsObject>,
  }

  export enum GroupStatus { 
    PENDING = 0,
    APPROVED = 1,
  }
}

export class Groups extends jspb.Message {
  getGroupsList(): Array<Group>;
  setGroupsList(value: Array<Group>): void;
  clearGroupsList(): void;
  addGroups(value?: Group, index?: number): Group;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Groups.AsObject;
  static toObject(includeInstance: boolean, msg: Groups): Groups.AsObject;
  static serializeBinaryToWriter(message: Groups, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Groups;
  static deserializeBinaryFromReader(message: Groups, reader: jspb.BinaryReader): Groups;
}

export namespace Groups {
  export type AsObject = {
    groupsList: Array<Group.AsObject>,
  }
}

export class Course extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getCoursecreatorid(): number;
  setCoursecreatorid(value: number): void;

  getName(): string;
  setName(value: string): void;

  getCode(): string;
  setCode(value: string): void;

  getYear(): number;
  setYear(value: number): void;

  getTag(): string;
  setTag(value: string): void;

  getProvider(): string;
  setProvider(value: string): void;

  getOrganizationid(): number;
  setOrganizationid(value: number): void;

  getOrganizationpath(): string;
  setOrganizationpath(value: string): void;

  getEnrolled(): Enrollment.UserStatus;
  setEnrolled(value: Enrollment.UserStatus): void;

  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): void;
  clearEnrollmentsList(): void;
  addEnrollments(value?: Enrollment, index?: number): Enrollment;

  getAssignmentsList(): Array<Assignment>;
  setAssignmentsList(value: Array<Assignment>): void;
  clearAssignmentsList(): void;
  addAssignments(value?: Assignment, index?: number): Assignment;

  getGroupsList(): Array<Group>;
  setGroupsList(value: Array<Group>): void;
  clearGroupsList(): void;
  addGroups(value?: Group, index?: number): Group;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Course.AsObject;
  static toObject(includeInstance: boolean, msg: Course): Course.AsObject;
  static serializeBinaryToWriter(message: Course, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Course;
  static deserializeBinaryFromReader(message: Course, reader: jspb.BinaryReader): Course;
}

export namespace Course {
  export type AsObject = {
    id: number,
    coursecreatorid: number,
    name: string,
    code: string,
    year: number,
    tag: string,
    provider: string,
    organizationid: number,
    organizationpath: string,
    enrolled: Enrollment.UserStatus,
    enrollmentsList: Array<Enrollment.AsObject>,
    assignmentsList: Array<Assignment.AsObject>,
    groupsList: Array<Group.AsObject>,
  }
}

export class Courses extends jspb.Message {
  getCoursesList(): Array<Course>;
  setCoursesList(value: Array<Course>): void;
  clearCoursesList(): void;
  addCourses(value?: Course, index?: number): Course;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Courses.AsObject;
  static toObject(includeInstance: boolean, msg: Courses): Courses.AsObject;
  static serializeBinaryToWriter(message: Courses, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Courses;
  static deserializeBinaryFromReader(message: Courses, reader: jspb.BinaryReader): Courses;
}

export namespace Courses {
  export type AsObject = {
    coursesList: Array<Course.AsObject>,
  }
}

export class Repository extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getOrganizationid(): number;
  setOrganizationid(value: number): void;

  getRepositoryid(): number;
  setRepositoryid(value: number): void;

  getUserid(): number;
  setUserid(value: number): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getHtmlurl(): string;
  setHtmlurl(value: string): void;

  getRepotype(): Repository.Type;
  setRepotype(value: Repository.Type): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Repository.AsObject;
  static toObject(includeInstance: boolean, msg: Repository): Repository.AsObject;
  static serializeBinaryToWriter(message: Repository, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Repository;
  static deserializeBinaryFromReader(message: Repository, reader: jspb.BinaryReader): Repository;
}

export namespace Repository {
  export type AsObject = {
    id: number,
    organizationid: number,
    repositoryid: number,
    userid: number,
    groupid: number,
    htmlurl: string,
    repotype: Repository.Type,
  }

  export enum Type { 
    NONE = 0,
    COURSEINFO = 1,
    ASSIGNMENTS = 2,
    TESTS = 3,
    SOLUTIONS = 4,
    USER = 5,
    GROUP = 6,
  }
}

export class Organization extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getPath(): string;
  setPath(value: string): void;

  getAvatar(): string;
  setAvatar(value: string): void;

  getPaymentplan(): string;
  setPaymentplan(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Organization.AsObject;
  static toObject(includeInstance: boolean, msg: Organization): Organization.AsObject;
  static serializeBinaryToWriter(message: Organization, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Organization;
  static deserializeBinaryFromReader(message: Organization, reader: jspb.BinaryReader): Organization;
}

export namespace Organization {
  export type AsObject = {
    id: number,
    path: string,
    avatar: string,
    paymentplan: string,
  }
}

export class Organizations extends jspb.Message {
  getOrganizationsList(): Array<Organization>;
  setOrganizationsList(value: Array<Organization>): void;
  clearOrganizationsList(): void;
  addOrganizations(value?: Organization, index?: number): Organization;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Organizations.AsObject;
  static toObject(includeInstance: boolean, msg: Organizations): Organizations.AsObject;
  static serializeBinaryToWriter(message: Organizations, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Organizations;
  static deserializeBinaryFromReader(message: Organizations, reader: jspb.BinaryReader): Organizations;
}

export namespace Organizations {
  export type AsObject = {
    organizationsList: Array<Organization.AsObject>,
  }
}

export class Enrollment extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  getUserid(): number;
  setUserid(value: number): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getHasteacherscopes(): boolean;
  setHasteacherscopes(value: boolean): void;

  getUser(): User | undefined;
  setUser(value?: User): void;
  hasUser(): boolean;
  clearUser(): void;

  getCourse(): Course | undefined;
  setCourse(value?: Course): void;
  hasCourse(): boolean;
  clearCourse(): void;

  getGroup(): Group | undefined;
  setGroup(value?: Group): void;
  hasGroup(): boolean;
  clearGroup(): void;

  getStatus(): Enrollment.UserStatus;
  setStatus(value: Enrollment.UserStatus): void;

  getState(): Enrollment.DisplayState;
  setState(value: Enrollment.DisplayState): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Enrollment.AsObject;
  static toObject(includeInstance: boolean, msg: Enrollment): Enrollment.AsObject;
  static serializeBinaryToWriter(message: Enrollment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Enrollment;
  static deserializeBinaryFromReader(message: Enrollment, reader: jspb.BinaryReader): Enrollment;
}

export namespace Enrollment {
  export type AsObject = {
    id: number,
    courseid: number,
    userid: number,
    groupid: number,
    hasteacherscopes: boolean,
    user?: User.AsObject,
    course?: Course.AsObject,
    group?: Group.AsObject,
    status: Enrollment.UserStatus,
    state: Enrollment.DisplayState,
  }

  export enum UserStatus { 
    NONE = 0,
    PENDING = 1,
    STUDENT = 2,
    TEACHER = 3,
  }

  export enum DisplayState { 
    UNSET = 0,
    HIDDEN = 1,
    VISIBLE = 2,
    FAVORITE = 3,
  }
}

export class Enrollments extends jspb.Message {
  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): void;
  clearEnrollmentsList(): void;
  addEnrollments(value?: Enrollment, index?: number): Enrollment;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Enrollments.AsObject;
  static toObject(includeInstance: boolean, msg: Enrollments): Enrollments.AsObject;
  static serializeBinaryToWriter(message: Enrollments, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Enrollments;
  static deserializeBinaryFromReader(message: Enrollments, reader: jspb.BinaryReader): Enrollments;
}

export namespace Enrollments {
  export type AsObject = {
    enrollmentsList: Array<Enrollment.AsObject>,
  }
}

export class Assignment extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  getName(): string;
  setName(value: string): void;

  getLanguage(): string;
  setLanguage(value: string): void;

  getDeadline(): string;
  setDeadline(value: string): void;

  getAutoapprove(): boolean;
  setAutoapprove(value: boolean): void;

  getOrder(): number;
  setOrder(value: number): void;

  getIsgrouplab(): boolean;
  setIsgrouplab(value: boolean): void;

  getScorelimit(): number;
  setScorelimit(value: number): void;

  getSubmissionsList(): Array<Submissions>;
  setSubmissionsList(value: Array<Submissions>): void;
  clearSubmissionsList(): void;
  addSubmissions(value?: Submissions, index?: number): Submissions;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Assignment.AsObject;
  static toObject(includeInstance: boolean, msg: Assignment): Assignment.AsObject;
  static serializeBinaryToWriter(message: Assignment, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Assignment;
  static deserializeBinaryFromReader(message: Assignment, reader: jspb.BinaryReader): Assignment;
}

export namespace Assignment {
  export type AsObject = {
    id: number,
    courseid: number,
    name: string,
    language: string,
    deadline: string,
    autoapprove: boolean,
    order: number,
    isgrouplab: boolean,
    scorelimit: number,
    submissionsList: Array<Submissions.AsObject>,
  }
}

export class Assignments extends jspb.Message {
  getAssignmentsList(): Array<Assignment>;
  setAssignmentsList(value: Array<Assignment>): void;
  clearAssignmentsList(): void;
  addAssignments(value?: Assignment, index?: number): Assignment;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Assignments.AsObject;
  static toObject(includeInstance: boolean, msg: Assignments): Assignments.AsObject;
  static serializeBinaryToWriter(message: Assignments, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Assignments;
  static deserializeBinaryFromReader(message: Assignments, reader: jspb.BinaryReader): Assignments;
}

export namespace Assignments {
  export type AsObject = {
    assignmentsList: Array<Assignment.AsObject>,
  }
}

export class Submission extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getAssignmentid(): number;
  setAssignmentid(value: number): void;

  getUserid(): number;
  setUserid(value: number): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getScore(): number;
  setScore(value: number): void;

  getScoreobjects(): string;
  setScoreobjects(value: string): void;

  getBuildinfo(): string;
  setBuildinfo(value: string): void;

  getCommithash(): string;
  setCommithash(value: string): void;

  getApproved(): boolean;
  setApproved(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Submission.AsObject;
  static toObject(includeInstance: boolean, msg: Submission): Submission.AsObject;
  static serializeBinaryToWriter(message: Submission, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Submission;
  static deserializeBinaryFromReader(message: Submission, reader: jspb.BinaryReader): Submission;
}

export namespace Submission {
  export type AsObject = {
    id: number,
    assignmentid: number,
    userid: number,
    groupid: number,
    score: number,
    scoreobjects: string,
    buildinfo: string,
    commithash: string,
    approved: boolean,
  }
}

export class Submissions extends jspb.Message {
  getSubmissionsList(): Array<Submission>;
  setSubmissionsList(value: Array<Submission>): void;
  clearSubmissionsList(): void;
  addSubmissions(value?: Submission, index?: number): Submission;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Submissions.AsObject;
  static toObject(includeInstance: boolean, msg: Submissions): Submissions.AsObject;
  static serializeBinaryToWriter(message: Submissions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Submissions;
  static deserializeBinaryFromReader(message: Submissions, reader: jspb.BinaryReader): Submissions;
}

export namespace Submissions {
  export type AsObject = {
    submissionsList: Array<Submission.AsObject>,
  }
}

export class SubmissionLink extends jspb.Message {
  getAssignment(): Assignment | undefined;
  setAssignment(value?: Assignment): void;
  hasAssignment(): boolean;
  clearAssignment(): void;

  getSubmission(): Submission | undefined;
  setSubmission(value?: Submission): void;
  hasSubmission(): boolean;
  clearSubmission(): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionLink.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionLink): SubmissionLink.AsObject;
  static serializeBinaryToWriter(message: SubmissionLink, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionLink;
  static deserializeBinaryFromReader(message: SubmissionLink, reader: jspb.BinaryReader): SubmissionLink;
}

export namespace SubmissionLink {
  export type AsObject = {
    assignment?: Assignment.AsObject,
    submission?: Submission.AsObject,
  }
}

export class AssignmentLink extends jspb.Message {
  getEnrollment(): Enrollment | undefined;
  setEnrollment(value?: Enrollment): void;
  hasEnrollment(): boolean;
  clearEnrollment(): void;

  getSubmissionsList(): Array<SubmissionLink>;
  setSubmissionsList(value: Array<SubmissionLink>): void;
  clearSubmissionsList(): void;
  addSubmissions(value?: SubmissionLink, index?: number): SubmissionLink;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AssignmentLink.AsObject;
  static toObject(includeInstance: boolean, msg: AssignmentLink): AssignmentLink.AsObject;
  static serializeBinaryToWriter(message: AssignmentLink, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AssignmentLink;
  static deserializeBinaryFromReader(message: AssignmentLink, reader: jspb.BinaryReader): AssignmentLink;
}

export namespace AssignmentLink {
  export type AsObject = {
    enrollment?: Enrollment.AsObject,
    submissionsList: Array<SubmissionLink.AsObject>,
  }
}

export class CourseSubmissions extends jspb.Message {
  getCourse(): Course | undefined;
  setCourse(value?: Course): void;
  hasCourse(): boolean;
  clearCourse(): void;

  getLinksList(): Array<AssignmentLink>;
  setLinksList(value: Array<AssignmentLink>): void;
  clearLinksList(): void;
  addLinks(value?: AssignmentLink, index?: number): AssignmentLink;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CourseSubmissions.AsObject;
  static toObject(includeInstance: boolean, msg: CourseSubmissions): CourseSubmissions.AsObject;
  static serializeBinaryToWriter(message: CourseSubmissions, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CourseSubmissions;
  static deserializeBinaryFromReader(message: CourseSubmissions, reader: jspb.BinaryReader): CourseSubmissions;
}

export namespace CourseSubmissions {
  export type AsObject = {
    course?: Course.AsObject,
    linksList: Array<AssignmentLink.AsObject>,
  }
}

export class CourseRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CourseRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CourseRequest): CourseRequest.AsObject;
  static serializeBinaryToWriter(message: CourseRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CourseRequest;
  static deserializeBinaryFromReader(message: CourseRequest, reader: jspb.BinaryReader): CourseRequest;
}

export namespace CourseRequest {
  export type AsObject = {
    courseid: number,
  }
}

export class UserRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UserRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UserRequest): UserRequest.AsObject;
  static serializeBinaryToWriter(message: UserRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UserRequest;
  static deserializeBinaryFromReader(message: UserRequest, reader: jspb.BinaryReader): UserRequest;
}

export namespace UserRequest {
  export type AsObject = {
    userid: number,
  }
}

export class GetGroupRequest extends jspb.Message {
  getGroupid(): number;
  setGroupid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetGroupRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetGroupRequest): GetGroupRequest.AsObject;
  static serializeBinaryToWriter(message: GetGroupRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetGroupRequest;
  static deserializeBinaryFromReader(message: GetGroupRequest, reader: jspb.BinaryReader): GetGroupRequest;
}

export namespace GetGroupRequest {
  export type AsObject = {
    groupid: number,
  }
}

export class GroupRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GroupRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GroupRequest): GroupRequest.AsObject;
  static serializeBinaryToWriter(message: GroupRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GroupRequest;
  static deserializeBinaryFromReader(message: GroupRequest, reader: jspb.BinaryReader): GroupRequest;
}

export namespace GroupRequest {
  export type AsObject = {
    userid: number,
    groupid: number,
    courseid: number,
  }
}

export class Provider extends jspb.Message {
  getProvider(): string;
  setProvider(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Provider.AsObject;
  static toObject(includeInstance: boolean, msg: Provider): Provider.AsObject;
  static serializeBinaryToWriter(message: Provider, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Provider;
  static deserializeBinaryFromReader(message: Provider, reader: jspb.BinaryReader): Provider;
}

export namespace Provider {
  export type AsObject = {
    provider: string,
  }
}

export class OrgRequest extends jspb.Message {
  getOrgname(): string;
  setOrgname(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OrgRequest.AsObject;
  static toObject(includeInstance: boolean, msg: OrgRequest): OrgRequest.AsObject;
  static serializeBinaryToWriter(message: OrgRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OrgRequest;
  static deserializeBinaryFromReader(message: OrgRequest, reader: jspb.BinaryReader): OrgRequest;
}

export namespace OrgRequest {
  export type AsObject = {
    orgname: string,
  }
}

export class EnrollmentRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): void;

  getIgnoregroupmembers(): boolean;
  setIgnoregroupmembers(value: boolean): void;

  getStatusesList(): Array<Enrollment.UserStatus>;
  setStatusesList(value: Array<Enrollment.UserStatus>): void;
  clearStatusesList(): void;
  addStatuses(value: Enrollment.UserStatus, index?: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnrollmentRequest.AsObject;
  static toObject(includeInstance: boolean, msg: EnrollmentRequest): EnrollmentRequest.AsObject;
  static serializeBinaryToWriter(message: EnrollmentRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnrollmentRequest;
  static deserializeBinaryFromReader(message: EnrollmentRequest, reader: jspb.BinaryReader): EnrollmentRequest;
}

export namespace EnrollmentRequest {
  export type AsObject = {
    courseid: number,
    ignoregroupmembers: boolean,
    statusesList: Array<Enrollment.UserStatus>,
  }
}

export class EnrollmentStatusRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): void;

  getStatusesList(): Array<Enrollment.UserStatus>;
  setStatusesList(value: Array<Enrollment.UserStatus>): void;
  clearStatusesList(): void;
  addStatuses(value: Enrollment.UserStatus, index?: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnrollmentStatusRequest.AsObject;
  static toObject(includeInstance: boolean, msg: EnrollmentStatusRequest): EnrollmentStatusRequest.AsObject;
  static serializeBinaryToWriter(message: EnrollmentStatusRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnrollmentStatusRequest;
  static deserializeBinaryFromReader(message: EnrollmentStatusRequest, reader: jspb.BinaryReader): EnrollmentStatusRequest;
}

export namespace EnrollmentStatusRequest {
  export type AsObject = {
    userid: number,
    statusesList: Array<Enrollment.UserStatus>,
  }
}

export class SubmissionRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionRequest): SubmissionRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionRequest;
  static deserializeBinaryFromReader(message: SubmissionRequest, reader: jspb.BinaryReader): SubmissionRequest;
}

export namespace SubmissionRequest {
  export type AsObject = {
    userid: number,
    groupid: number,
    courseid: number,
  }
}

export class UpdateSubmissionRequest extends jspb.Message {
  getSubmissionid(): number;
  setSubmissionid(value: number): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  getApprove(): boolean;
  setApprove(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateSubmissionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateSubmissionRequest): UpdateSubmissionRequest.AsObject;
  static serializeBinaryToWriter(message: UpdateSubmissionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateSubmissionRequest;
  static deserializeBinaryFromReader(message: UpdateSubmissionRequest, reader: jspb.BinaryReader): UpdateSubmissionRequest;
}

export namespace UpdateSubmissionRequest {
  export type AsObject = {
    submissionid: number,
    courseid: number,
    approve: boolean,
  }
}

export class Providers extends jspb.Message {
  getProvidersList(): Array<string>;
  setProvidersList(value: Array<string>): void;
  clearProvidersList(): void;
  addProviders(value: string, index?: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Providers.AsObject;
  static toObject(includeInstance: boolean, msg: Providers): Providers.AsObject;
  static serializeBinaryToWriter(message: Providers, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Providers;
  static deserializeBinaryFromReader(message: Providers, reader: jspb.BinaryReader): Providers;
}

export namespace Providers {
  export type AsObject = {
    providersList: Array<string>,
  }
}

export class URLRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): void;

  getRepotypesList(): Array<Repository.Type>;
  setRepotypesList(value: Array<Repository.Type>): void;
  clearRepotypesList(): void;
  addRepotypes(value: Repository.Type, index?: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): URLRequest.AsObject;
  static toObject(includeInstance: boolean, msg: URLRequest): URLRequest.AsObject;
  static serializeBinaryToWriter(message: URLRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): URLRequest;
  static deserializeBinaryFromReader(message: URLRequest, reader: jspb.BinaryReader): URLRequest;
}

export namespace URLRequest {
  export type AsObject = {
    courseid: number,
    repotypesList: Array<Repository.Type>,
  }
}

export class RepositoryRequest extends jspb.Message {
  getUserid(): number;
  setUserid(value: number): void;

  getGroupid(): number;
  setGroupid(value: number): void;

  getCourseid(): number;
  setCourseid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RepositoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RepositoryRequest): RepositoryRequest.AsObject;
  static serializeBinaryToWriter(message: RepositoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RepositoryRequest;
  static deserializeBinaryFromReader(message: RepositoryRequest, reader: jspb.BinaryReader): RepositoryRequest;
}

export namespace RepositoryRequest {
  export type AsObject = {
    userid: number,
    groupid: number,
    courseid: number,
  }
}

export class Repositories extends jspb.Message {
  getUrlsMap(): jspb.Map<string, string>;
  clearUrlsMap(): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Repositories.AsObject;
  static toObject(includeInstance: boolean, msg: Repositories): Repositories.AsObject;
  static serializeBinaryToWriter(message: Repositories, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Repositories;
  static deserializeBinaryFromReader(message: Repositories, reader: jspb.BinaryReader): Repositories;
}

export namespace Repositories {
  export type AsObject = {
    urlsMap: Array<[string, string]>,
  }
}

export class AuthorizationResponse extends jspb.Message {
  getIsauthorized(): boolean;
  setIsauthorized(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AuthorizationResponse.AsObject;
  static toObject(includeInstance: boolean, msg: AuthorizationResponse): AuthorizationResponse.AsObject;
  static serializeBinaryToWriter(message: AuthorizationResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AuthorizationResponse;
  static deserializeBinaryFromReader(message: AuthorizationResponse, reader: jspb.BinaryReader): AuthorizationResponse;
}

export namespace AuthorizationResponse {
  export type AsObject = {
    isauthorized: boolean,
  }
}

export class Status extends jspb.Message {
  getCode(): number;
  setCode(value: number): void;

  getError(): string;
  setError(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Status.AsObject;
  static toObject(includeInstance: boolean, msg: Status): Status.AsObject;
  static serializeBinaryToWriter(message: Status, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Status;
  static deserializeBinaryFromReader(message: Status, reader: jspb.BinaryReader): Status;
}

export namespace Status {
  export type AsObject = {
    code: number,
    error: string,
  }
}

export class SubmissionLinkRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): void;

  getType(): SubmissionLinkRequest.Type;
  setType(value: SubmissionLinkRequest.Type): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubmissionLinkRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubmissionLinkRequest): SubmissionLinkRequest.AsObject;
  static serializeBinaryToWriter(message: SubmissionLinkRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubmissionLinkRequest;
  static deserializeBinaryFromReader(message: SubmissionLinkRequest, reader: jspb.BinaryReader): SubmissionLinkRequest;
}

export namespace SubmissionLinkRequest {
  export type AsObject = {
    courseid: number,
    type: SubmissionLinkRequest.Type,
  }

  export enum Type { 
    ALL = 0,
    INDIVIDUAL = 1,
    GROUP = 2,
  }
}

export class RebuildRequest extends jspb.Message {
  getCourseid(): number;
  setCourseid(value: number): void;

  getSubmissionid(): number;
  setSubmissionid(value: number): void;

  getAssignmentid(): number;
  setAssignmentid(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RebuildRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RebuildRequest): RebuildRequest.AsObject;
  static serializeBinaryToWriter(message: RebuildRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RebuildRequest;
  static deserializeBinaryFromReader(message: RebuildRequest, reader: jspb.BinaryReader): RebuildRequest;
}

export namespace RebuildRequest {
  export type AsObject = {
    courseid: number,
    submissionid: number,
    assignmentid: number,
  }
}

export class Void extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Void.AsObject;
  static toObject(includeInstance: boolean, msg: Void): Void.AsObject;
  static serializeBinaryToWriter(message: Void, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Void;
  static deserializeBinaryFromReader(message: Void, reader: jspb.BinaryReader): Void;
}

export namespace Void {
  export type AsObject = {
  }
}

