import * as jspb from 'google-protobuf'

import * as google_protobuf_timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';
import * as kit_score_score_pb from '../kit/score/score_pb';


export class User extends jspb.Message {
  getId(): number;
  setId(value: number): User;

  getIsadmin(): boolean;
  setIsadmin(value: boolean): User;

  getName(): string;
  setName(value: string): User;

  getStudentid(): string;
  setStudentid(value: string): User;

  getEmail(): string;
  setEmail(value: string): User;

  getAvatarurl(): string;
  setAvatarurl(value: string): User;

  getLogin(): string;
  setLogin(value: string): User;

  getRemoteidentitiesList(): Array<RemoteIdentity>;
  setRemoteidentitiesList(value: Array<RemoteIdentity>): User;
  clearRemoteidentitiesList(): User;
  addRemoteidentities(value?: RemoteIdentity, index?: number): RemoteIdentity;

  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): User;
  clearEnrollmentsList(): User;
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
  setUsersList(value: Array<User>): Users;
  clearUsersList(): Users;
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
  setId(value: number): RemoteIdentity;

  getProvider(): string;
  setProvider(value: string): RemoteIdentity;

  getRemoteid(): number;
  setRemoteid(value: number): RemoteIdentity;

  getAccesstoken(): string;
  setAccesstoken(value: string): RemoteIdentity;

  getUserid(): number;
  setUserid(value: number): RemoteIdentity;

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
  setId(value: number): Group;

  getName(): string;
  setName(value: string): Group;

  getCourseid(): number;
  setCourseid(value: number): Group;

  getTeamid(): number;
  setTeamid(value: number): Group;

  getStatus(): Group.GroupStatus;
  setStatus(value: Group.GroupStatus): Group;

  getUsersList(): Array<User>;
  setUsersList(value: Array<User>): Group;
  clearUsersList(): Group;
  addUsers(value?: User, index?: number): User;

  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): Group;
  clearEnrollmentsList(): Group;
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
  setGroupsList(value: Array<Group>): Groups;
  clearGroupsList(): Groups;
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
  setId(value: number): Course;

  getCoursecreatorid(): number;
  setCoursecreatorid(value: number): Course;

  getName(): string;
  setName(value: string): Course;

  getCode(): string;
  setCode(value: string): Course;

  getYear(): number;
  setYear(value: number): Course;

  getTag(): string;
  setTag(value: string): Course;

  getProvider(): string;
  setProvider(value: string): Course;

  getOrganizationid(): number;
  setOrganizationid(value: number): Course;

  getOrganizationpath(): string;
  setOrganizationpath(value: string): Course;

  getSlipdays(): number;
  setSlipdays(value: number): Course;

  getDockerfile(): string;
  setDockerfile(value: string): Course;

  getEnrolled(): Enrollment.UserStatus;
  setEnrolled(value: Enrollment.UserStatus): Course;

  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): Course;
  clearEnrollmentsList(): Course;
  addEnrollments(value?: Enrollment, index?: number): Enrollment;

  getAssignmentsList(): Array<Assignment>;
  setAssignmentsList(value: Array<Assignment>): Course;
  clearAssignmentsList(): Course;
  addAssignments(value?: Assignment, index?: number): Assignment;

  getGroupsList(): Array<Group>;
  setGroupsList(value: Array<Group>): Course;
  clearGroupsList(): Course;
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
    slipdays: number,
    dockerfile: string,
    enrolled: Enrollment.UserStatus,
    enrollmentsList: Array<Enrollment.AsObject>,
    assignmentsList: Array<Assignment.AsObject>,
    groupsList: Array<Group.AsObject>,
  }
}

export class Courses extends jspb.Message {
  getCoursesList(): Array<Course>;
  setCoursesList(value: Array<Course>): Courses;
  clearCoursesList(): Courses;
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
  setId(value: number): Repository;

  getOrganizationid(): number;
  setOrganizationid(value: number): Repository;

  getRepositoryid(): number;
  setRepositoryid(value: number): Repository;

  getUserid(): number;
  setUserid(value: number): Repository;

  getGroupid(): number;
  setGroupid(value: number): Repository;

  getHtmlurl(): string;
  setHtmlurl(value: string): Repository;

  getRepotype(): Repository.Type;
  setRepotype(value: Repository.Type): Repository;

  getIssuesList(): Array<Issue>;
  setIssuesList(value: Array<Issue>): Repository;
  clearIssuesList(): Repository;
  addIssues(value?: Issue, index?: number): Issue;

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
    issuesList: Array<Issue.AsObject>,
  }

  export enum Type { 
    NONE = 0,
    INFO = 1,
    ASSIGNMENTS = 2,
    TESTS = 3,
    USER = 4,
    GROUP = 5,
  }
}

export class Enrollment extends jspb.Message {
  getId(): number;
  setId(value: number): Enrollment;

  getCourseid(): number;
  setCourseid(value: number): Enrollment;

  getUserid(): number;
  setUserid(value: number): Enrollment;

  getGroupid(): number;
  setGroupid(value: number): Enrollment;

  getHasteacherscopes(): boolean;
  setHasteacherscopes(value: boolean): Enrollment;

  getUser(): User | undefined;
  setUser(value?: User): Enrollment;
  hasUser(): boolean;
  clearUser(): Enrollment;

  getCourse(): Course | undefined;
  setCourse(value?: Course): Enrollment;
  hasCourse(): boolean;
  clearCourse(): Enrollment;

  getGroup(): Group | undefined;
  setGroup(value?: Group): Enrollment;
  hasGroup(): boolean;
  clearGroup(): Enrollment;

  getStatus(): Enrollment.UserStatus;
  setStatus(value: Enrollment.UserStatus): Enrollment;

  getState(): Enrollment.DisplayState;
  setState(value: Enrollment.DisplayState): Enrollment;

  getSlipdaysremaining(): number;
  setSlipdaysremaining(value: number): Enrollment;

  getLastactivitydate(): string;
  setLastactivitydate(value: string): Enrollment;

  getTotalapproved(): number;
  setTotalapproved(value: number): Enrollment;

  getUsedslipdaysList(): Array<UsedSlipDays>;
  setUsedslipdaysList(value: Array<UsedSlipDays>): Enrollment;
  clearUsedslipdaysList(): Enrollment;
  addUsedslipdays(value?: UsedSlipDays, index?: number): UsedSlipDays;

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
    slipdaysremaining: number,
    lastactivitydate: string,
    totalapproved: number,
    usedslipdaysList: Array<UsedSlipDays.AsObject>,
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

export class UsedSlipDays extends jspb.Message {
  getId(): number;
  setId(value: number): UsedSlipDays;

  getEnrollmentid(): number;
  setEnrollmentid(value: number): UsedSlipDays;

  getAssignmentid(): number;
  setAssignmentid(value: number): UsedSlipDays;

  getUseddays(): number;
  setUseddays(value: number): UsedSlipDays;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UsedSlipDays.AsObject;
  static toObject(includeInstance: boolean, msg: UsedSlipDays): UsedSlipDays.AsObject;
  static serializeBinaryToWriter(message: UsedSlipDays, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UsedSlipDays;
  static deserializeBinaryFromReader(message: UsedSlipDays, reader: jspb.BinaryReader): UsedSlipDays;
}

export namespace UsedSlipDays {
  export type AsObject = {
    id: number,
    enrollmentid: number,
    assignmentid: number,
    useddays: number,
  }
}

export class Enrollments extends jspb.Message {
  getEnrollmentsList(): Array<Enrollment>;
  setEnrollmentsList(value: Array<Enrollment>): Enrollments;
  clearEnrollmentsList(): Enrollments;
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
  setId(value: number): Assignment;

  getCourseid(): number;
  setCourseid(value: number): Assignment;

  getName(): string;
  setName(value: string): Assignment;

  getRunscriptcontent(): string;
  setRunscriptcontent(value: string): Assignment;

  getDeadline(): string;
  setDeadline(value: string): Assignment;

  getAutoapprove(): boolean;
  setAutoapprove(value: boolean): Assignment;

  getOrder(): number;
  setOrder(value: number): Assignment;

  getIsgrouplab(): boolean;
  setIsgrouplab(value: boolean): Assignment;

  getScorelimit(): number;
  setScorelimit(value: number): Assignment;

  getReviewers(): number;
  setReviewers(value: number): Assignment;

  getContainertimeout(): number;
  setContainertimeout(value: number): Assignment;

  getSubmissionsList(): Array<Submission>;
  setSubmissionsList(value: Array<Submission>): Assignment;
  clearSubmissionsList(): Assignment;
  addSubmissions(value?: Submission, index?: number): Submission;

  getTasksList(): Array<Task>;
  setTasksList(value: Array<Task>): Assignment;
  clearTasksList(): Assignment;
  addTasks(value?: Task, index?: number): Task;

  getGradingbenchmarksList(): Array<GradingBenchmark>;
  setGradingbenchmarksList(value: Array<GradingBenchmark>): Assignment;
  clearGradingbenchmarksList(): Assignment;
  addGradingbenchmarks(value?: GradingBenchmark, index?: number): GradingBenchmark;

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
    runscriptcontent: string,
    deadline: string,
    autoapprove: boolean,
    order: number,
    isgrouplab: boolean,
    scorelimit: number,
    reviewers: number,
    containertimeout: number,
    submissionsList: Array<Submission.AsObject>,
    tasksList: Array<Task.AsObject>,
    gradingbenchmarksList: Array<GradingBenchmark.AsObject>,
  }
}

export class Task extends jspb.Message {
  getId(): number;
  setId(value: number): Task;

  getAssignmentid(): number;
  setAssignmentid(value: number): Task;

  getAssignmentorder(): number;
  setAssignmentorder(value: number): Task;

  getTitle(): string;
  setTitle(value: string): Task;

  getBody(): string;
  setBody(value: string): Task;

  getName(): string;
  setName(value: string): Task;

  getIssuesList(): Array<Issue>;
  setIssuesList(value: Array<Issue>): Task;
  clearIssuesList(): Task;
  addIssues(value?: Issue, index?: number): Issue;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Task.AsObject;
  static toObject(includeInstance: boolean, msg: Task): Task.AsObject;
  static serializeBinaryToWriter(message: Task, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Task;
  static deserializeBinaryFromReader(message: Task, reader: jspb.BinaryReader): Task;
}

export namespace Task {
  export type AsObject = {
    id: number,
    assignmentid: number,
    assignmentorder: number,
    title: string,
    body: string,
    name: string,
    issuesList: Array<Issue.AsObject>,
  }
}

export class Issue extends jspb.Message {
  getId(): number;
  setId(value: number): Issue;

  getRepositoryid(): number;
  setRepositoryid(value: number): Issue;

  getTaskid(): number;
  setTaskid(value: number): Issue;

  getIssuenumber(): number;
  setIssuenumber(value: number): Issue;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Issue.AsObject;
  static toObject(includeInstance: boolean, msg: Issue): Issue.AsObject;
  static serializeBinaryToWriter(message: Issue, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Issue;
  static deserializeBinaryFromReader(message: Issue, reader: jspb.BinaryReader): Issue;
}

export namespace Issue {
  export type AsObject = {
    id: number,
    repositoryid: number,
    taskid: number,
    issuenumber: number,
  }
}

export class PullRequest extends jspb.Message {
  getId(): number;
  setId(value: number): PullRequest;

  getScmrepositoryid(): number;
  setScmrepositoryid(value: number): PullRequest;

  getTaskid(): number;
  setTaskid(value: number): PullRequest;

  getIssueid(): number;
  setIssueid(value: number): PullRequest;

  getUserid(): number;
  setUserid(value: number): PullRequest;

  getScmcommentid(): number;
  setScmcommentid(value: number): PullRequest;

  getSourcebranch(): string;
  setSourcebranch(value: string): PullRequest;

  getNumber(): number;
  setNumber(value: number): PullRequest;

  getStage(): PullRequest.Stage;
  setStage(value: PullRequest.Stage): PullRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PullRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PullRequest): PullRequest.AsObject;
  static serializeBinaryToWriter(message: PullRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PullRequest;
  static deserializeBinaryFromReader(message: PullRequest, reader: jspb.BinaryReader): PullRequest;
}

export namespace PullRequest {
  export type AsObject = {
    id: number,
    scmrepositoryid: number,
    taskid: number,
    issueid: number,
    userid: number,
    scmcommentid: number,
    sourcebranch: string,
    number: number,
    stage: PullRequest.Stage,
  }

  export enum Stage { 
    NONE = 0,
    DRAFT = 1,
    REVIEW = 2,
    APPROVED = 3,
  }
}

export class Assignments extends jspb.Message {
  getAssignmentsList(): Array<Assignment>;
  setAssignmentsList(value: Array<Assignment>): Assignments;
  clearAssignmentsList(): Assignments;
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
  setId(value: number): Submission;

  getAssignmentid(): number;
  setAssignmentid(value: number): Submission;

  getUserid(): number;
  setUserid(value: number): Submission;

  getGroupid(): number;
  setGroupid(value: number): Submission;

  getScore(): number;
  setScore(value: number): Submission;

  getCommithash(): string;
  setCommithash(value: string): Submission;

  getReleased(): boolean;
  setReleased(value: boolean): Submission;

  getStatus(): Submission.Status;
  setStatus(value: Submission.Status): Submission;

  getApproveddate(): google_protobuf_timestamp_pb.Timestamp | undefined;
  setApproveddate(value?: google_protobuf_timestamp_pb.Timestamp): Submission;
  hasApproveddate(): boolean;
  clearApproveddate(): Submission;

  getReviewsList(): Array<Review>;
  setReviewsList(value: Array<Review>): Submission;
  clearReviewsList(): Submission;
  addReviews(value?: Review, index?: number): Review;

  getBuildinfo(): kit_score_score_pb.BuildInfo | undefined;
  setBuildinfo(value?: kit_score_score_pb.BuildInfo): Submission;
  hasBuildinfo(): boolean;
  clearBuildinfo(): Submission;

  getScoresList(): Array<kit_score_score_pb.Score>;
  setScoresList(value: Array<kit_score_score_pb.Score>): Submission;
  clearScoresList(): Submission;
  addScores(value?: kit_score_score_pb.Score, index?: number): kit_score_score_pb.Score;

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
    commithash: string,
    released: boolean,
    status: Submission.Status,
    approveddate?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    reviewsList: Array<Review.AsObject>,
    buildinfo?: kit_score_score_pb.BuildInfo.AsObject,
    scoresList: Array<kit_score_score_pb.Score.AsObject>,
  }

  export enum Status { 
    NONE = 0,
    APPROVED = 1,
    REJECTED = 2,
    REVISION = 3,
  }
}

export class Submissions extends jspb.Message {
  getSubmissionsList(): Array<Submission>;
  setSubmissionsList(value: Array<Submission>): Submissions;
  clearSubmissionsList(): Submissions;
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

export class GradingBenchmark extends jspb.Message {
  getId(): number;
  setId(value: number): GradingBenchmark;

  getAssignmentid(): number;
  setAssignmentid(value: number): GradingBenchmark;

  getReviewid(): number;
  setReviewid(value: number): GradingBenchmark;

  getHeading(): string;
  setHeading(value: string): GradingBenchmark;

  getComment(): string;
  setComment(value: string): GradingBenchmark;

  getCriteriaList(): Array<GradingCriterion>;
  setCriteriaList(value: Array<GradingCriterion>): GradingBenchmark;
  clearCriteriaList(): GradingBenchmark;
  addCriteria(value?: GradingCriterion, index?: number): GradingCriterion;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GradingBenchmark.AsObject;
  static toObject(includeInstance: boolean, msg: GradingBenchmark): GradingBenchmark.AsObject;
  static serializeBinaryToWriter(message: GradingBenchmark, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GradingBenchmark;
  static deserializeBinaryFromReader(message: GradingBenchmark, reader: jspb.BinaryReader): GradingBenchmark;
}

export namespace GradingBenchmark {
  export type AsObject = {
    id: number,
    assignmentid: number,
    reviewid: number,
    heading: string,
    comment: string,
    criteriaList: Array<GradingCriterion.AsObject>,
  }
}

export class Benchmarks extends jspb.Message {
  getBenchmarksList(): Array<GradingBenchmark>;
  setBenchmarksList(value: Array<GradingBenchmark>): Benchmarks;
  clearBenchmarksList(): Benchmarks;
  addBenchmarks(value?: GradingBenchmark, index?: number): GradingBenchmark;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Benchmarks.AsObject;
  static toObject(includeInstance: boolean, msg: Benchmarks): Benchmarks.AsObject;
  static serializeBinaryToWriter(message: Benchmarks, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Benchmarks;
  static deserializeBinaryFromReader(message: Benchmarks, reader: jspb.BinaryReader): Benchmarks;
}

export namespace Benchmarks {
  export type AsObject = {
    benchmarksList: Array<GradingBenchmark.AsObject>,
  }
}

export class GradingCriterion extends jspb.Message {
  getId(): number;
  setId(value: number): GradingCriterion;

  getBenchmarkid(): number;
  setBenchmarkid(value: number): GradingCriterion;

  getPoints(): number;
  setPoints(value: number): GradingCriterion;

  getDescription(): string;
  setDescription(value: string): GradingCriterion;

  getGrade(): GradingCriterion.Grade;
  setGrade(value: GradingCriterion.Grade): GradingCriterion;

  getComment(): string;
  setComment(value: string): GradingCriterion;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GradingCriterion.AsObject;
  static toObject(includeInstance: boolean, msg: GradingCriterion): GradingCriterion.AsObject;
  static serializeBinaryToWriter(message: GradingCriterion, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GradingCriterion;
  static deserializeBinaryFromReader(message: GradingCriterion, reader: jspb.BinaryReader): GradingCriterion;
}

export namespace GradingCriterion {
  export type AsObject = {
    id: number,
    benchmarkid: number,
    points: number,
    description: string,
    grade: GradingCriterion.Grade,
    comment: string,
  }

  export enum Grade { 
    NONE = 0,
    FAILED = 1,
    PASSED = 2,
  }
}

export class Review extends jspb.Message {
  getId(): number;
  setId(value: number): Review;

  getSubmissionid(): number;
  setSubmissionid(value: number): Review;

  getReviewerid(): number;
  setReviewerid(value: number): Review;

  getFeedback(): string;
  setFeedback(value: string): Review;

  getReady(): boolean;
  setReady(value: boolean): Review;

  getScore(): number;
  setScore(value: number): Review;

  getGradingbenchmarksList(): Array<GradingBenchmark>;
  setGradingbenchmarksList(value: Array<GradingBenchmark>): Review;
  clearGradingbenchmarksList(): Review;
  addGradingbenchmarks(value?: GradingBenchmark, index?: number): GradingBenchmark;

  getEdited(): string;
  setEdited(value: string): Review;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Review.AsObject;
  static toObject(includeInstance: boolean, msg: Review): Review.AsObject;
  static serializeBinaryToWriter(message: Review, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Review;
  static deserializeBinaryFromReader(message: Review, reader: jspb.BinaryReader): Review;
}

export namespace Review {
  export type AsObject = {
    id: number,
    submissionid: number,
    reviewerid: number,
    feedback: string,
    ready: boolean,
    score: number,
    gradingbenchmarksList: Array<GradingBenchmark.AsObject>,
    edited: string,
  }
}

export class SubmissionLink extends jspb.Message {
  getAssignment(): Assignment | undefined;
  setAssignment(value?: Assignment): SubmissionLink;
  hasAssignment(): boolean;
  clearAssignment(): SubmissionLink;

  getSubmission(): Submission | undefined;
  setSubmission(value?: Submission): SubmissionLink;
  hasSubmission(): boolean;
  clearSubmission(): SubmissionLink;

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

export class EnrollmentLink extends jspb.Message {
  getEnrollment(): Enrollment | undefined;
  setEnrollment(value?: Enrollment): EnrollmentLink;
  hasEnrollment(): boolean;
  clearEnrollment(): EnrollmentLink;

  getSubmissionsList(): Array<SubmissionLink>;
  setSubmissionsList(value: Array<SubmissionLink>): EnrollmentLink;
  clearSubmissionsList(): EnrollmentLink;
  addSubmissions(value?: SubmissionLink, index?: number): SubmissionLink;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EnrollmentLink.AsObject;
  static toObject(includeInstance: boolean, msg: EnrollmentLink): EnrollmentLink.AsObject;
  static serializeBinaryToWriter(message: EnrollmentLink, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EnrollmentLink;
  static deserializeBinaryFromReader(message: EnrollmentLink, reader: jspb.BinaryReader): EnrollmentLink;
}

export namespace EnrollmentLink {
  export type AsObject = {
    enrollment?: Enrollment.AsObject,
    submissionsList: Array<SubmissionLink.AsObject>,
  }
}

