export class ActionRequest {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getUserId(): number;
  setUserId(a: number): void;
  getGroupId(): number;
  setGroupId(a: number): void;
  getCourseId(): number;
  setCourseId(a: number): void;
  getStatus(): Enrollment.UserStatus;
  setStatus(a: Enrollment.UserStatus): void;
  getGroupStatus(): Group.GroupStatus;
  setGroupStatus(a: Group.GroupStatus): void;
  toObject(): ActionRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => ActionRequest;
}

export namespace ActionRequest {
  export type AsObject = {
    Id: number;
    UserId: number;
    GroupId: number;
    CourseId: number;
    Status: Enrollment.UserStatus;
    GroupStatus: Group.GroupStatus;
  }
}

export class Assignment {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getCourseId(): number;
  setCourseId(a: number): void;
  getName(): string;
  setName(a: string): void;
  getLanguage(): string;
  setLanguage(a: string): void;
  getDeadline(): Timestamp;
  setDeadline(a: Timestamp): void;
  getAutoApprove(): boolean;
  setAutoApprove(a: boolean): void;
  getOrder(): number;
  setOrder(a: number): void;
  getIsGrouplab(): boolean;
  setIsGrouplab(a: boolean): void;
  getSubmission(): Submission;
  setSubmission(a: Submission): void;
  toObject(): Assignment.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Assignment;
}

export namespace Assignment {
  export type AsObject = {
    Id: number;
    CourseId: number;
    Name: string;
    Language: string;
    Deadline: Timestamp;
    AutoApprove: boolean;
    Order: number;
    IsGrouplab: boolean;
    Submission: Submission;
  }
}

export class Assignments {
  constructor ();
  getAssignmentsList(): Assignment[];
  setAssignmentsList(a: Assignment[]): void;
  toObject(): Assignments.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Assignments;
}

export namespace Assignments {
  export type AsObject = {
    AssignmentsList: Assignment[];
  }
}

export class Course {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getCoursecreatorId(): number;
  setCoursecreatorId(a: number): void;
  getName(): string;
  setName(a: string): void;
  getCode(): string;
  setCode(a: string): void;
  getYear(): number;
  setYear(a: number): void;
  getTag(): string;
  setTag(a: string): void;
  getProvider(): string;
  setProvider(a: string): void;
  getDirectoryId(): number;
  setDirectoryId(a: number): void;
  getEnrolled(): Enrollment.UserStatus;
  setEnrolled(a: Enrollment.UserStatus): void;
  getEnrollmentsList(): Enrollment[];
  setEnrollmentsList(a: Enrollment[]): void;
  getAssignmentsList(): Assignment[];
  setAssignmentsList(a: Assignment[]): void;
  getGroupsList(): Group[];
  setGroupsList(a: Group[]): void;
  toObject(): Course.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Course;
}

export namespace Course {
  export type AsObject = {
    Id: number;
    CoursecreatorId: number;
    Name: string;
    Code: string;
    Year: number;
    Tag: string;
    Provider: string;
    DirectoryId: number;
    Enrolled: Enrollment.UserStatus;
    EnrollmentsList: Enrollment[];
    AssignmentsList: Assignment[];
    GroupsList: Group[];
  }
}

export class Courses {
  constructor ();
  getCoursesList(): Course[];
  setCoursesList(a: Course[]): void;
  toObject(): Courses.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Courses;
}

export namespace Courses {
  export type AsObject = {
    CoursesList: Course[];
  }
}

export class Directories {
  constructor ();
  getDirectoriesList(): Directory[];
  setDirectoriesList(a: Directory[]): void;
  toObject(): Directories.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Directories;
}

export namespace Directories {
  export type AsObject = {
    DirectoriesList: Directory[];
  }
}

export class Directory {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getPath(): string;
  setPath(a: string): void;
  getAvatar(): string;
  setAvatar(a: string): void;
  toObject(): Directory.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Directory;
}

export namespace Directory {
  export type AsObject = {
    Id: number;
    Path: string;
    Avatar: string;
  }
}

export class DirectoryRequest {
  constructor ();
  getProvider(): string;
  setProvider(a: string): void;
  getCourseId(): number;
  setCourseId(a: number): void;
  toObject(): DirectoryRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => DirectoryRequest;
}

export namespace DirectoryRequest {
  export type AsObject = {
    Provider: string;
    CourseId: number;
  }
}

export class Enrollment {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getCourseId(): number;
  setCourseId(a: number): void;
  getUserId(): number;
  setUserId(a: number): void;
  getGroupId(): number;
  setGroupId(a: number): void;
  getUser(): User;
  setUser(a: User): void;
  getCourse(): Course;
  setCourse(a: Course): void;
  getGroup(): Group;
  setGroup(a: Group): void;
  getStatus(): Enrollment.UserStatus;
  setStatus(a: Enrollment.UserStatus): void;
  toObject(): Enrollment.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Enrollment;
}

export namespace Enrollment {
  export type AsObject = {
    Id: number;
    CourseId: number;
    UserId: number;
    GroupId: number;
    User: User;
    Course: Course;
    Group: Group;
    Status: Enrollment.UserStatus;
  }

  export enum UserStatus { 
    PENDING = 0,
    REJECTED = 1,
    STUDENT = 2,
    TEACHER = 3,
  }
}

export class Enrollments {
  constructor ();
  getEnrollmentsList(): Enrollment[];
  setEnrollmentsList(a: Enrollment[]): void;
  toObject(): Enrollments.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Enrollments;
}

export namespace Enrollments {
  export type AsObject = {
    EnrollmentsList: Enrollment[];
  }
}

export class Group {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getName(): string;
  setName(a: string): void;
  getCourseId(): number;
  setCourseId(a: number): void;
  getStatus(): Group.GroupStatus;
  setStatus(a: Group.GroupStatus): void;
  getUsersList(): User[];
  setUsersList(a: User[]): void;
  getEnrollmentsList(): Enrollment[];
  setEnrollmentsList(a: Enrollment[]): void;
  toObject(): Group.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Group;
}

export namespace Group {
  export type AsObject = {
    Id: number;
    Name: string;
    CourseId: number;
    Status: Group.GroupStatus;
    UsersList: User[];
    EnrollmentsList: Enrollment[];
  }

  export enum GroupStatus { 
    PENDING_GROUP = 0,
    REJECTED_GROUP = 1,
    APPROVED = 2,
    DELETED = 3,
  }
}

export class Groups {
  constructor ();
  getGroupsList(): Group[];
  setGroupsList(a: Group[]): void;
  toObject(): Groups.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Groups;
}

export namespace Groups {
  export type AsObject = {
    GroupsList: Group[];
  }
}

export class Providers {
  constructor ();
  getProvidersList(): string[];
  setProvidersList(a: string[]): void;
  toObject(): Providers.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Providers;
}

export namespace Providers {
  export type AsObject = {
    ProvidersList: string[];
  }
}

export class RecordRequest {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getStatusesList(): Enrollment.UserStatus[];
  setStatusesList(a: Enrollment.UserStatus[]): void;
  getGroupStatusesList(): Group.GroupStatus[];
  setGroupStatusesList(a: Group.GroupStatus[]): void;
  toObject(): RecordRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => RecordRequest;
}

export namespace RecordRequest {
  export type AsObject = {
    Id: number;
    StatusesList: Enrollment.UserStatus[];
    GroupStatusesList: Group.GroupStatus[];
  }
}

export class RemoteIdentity {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getProvider(): string;
  setProvider(a: string): void;
  getRemoteId(): number;
  setRemoteId(a: number): void;
  getAccessToken(): string;
  setAccessToken(a: string): void;
  getUserId(): number;
  setUserId(a: number): void;
  toObject(): RemoteIdentity.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => RemoteIdentity;
}

export namespace RemoteIdentity {
  export type AsObject = {
    Id: number;
    Provider: string;
    RemoteId: number;
    AccessToken: string;
    UserId: number;
  }
}

export class Repository {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getDirectoryId(): number;
  setDirectoryId(a: number): void;
  getRepositoryId(): number;
  setRepositoryId(a: number): void;
  getUserId(): number;
  setUserId(a: number): void;
  getGroupId(): number;
  setGroupId(a: number): void;
  getHtmlUrl(): string;
  setHtmlUrl(a: string): void;
  getRepoType(): Repository.RepoType;
  setRepoType(a: Repository.RepoType): void;
  toObject(): Repository.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Repository;
}

export namespace Repository {
  export type AsObject = {
    Id: number;
    DirectoryId: number;
    RepositoryId: number;
    UserId: number;
    GroupId: number;
    HtmlUrl: string;
    RepoType: Repository.RepoType;
  }

  export enum RepoType { 
    USER = 0,
    ASSIGNMENT = 1,
    TESTS = 2,
    SOLUTION = 3,
    COURSEINFO = 4,
  }
}

export class RepositoryRequest {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getType(): Repository.RepoType;
  setType(a: Repository.RepoType): void;
  getDirectoryId(): number;
  setDirectoryId(a: number): void;
  getRepositoryId(): number;
  setRepositoryId(a: number): void;
  getUserId(): number;
  setUserId(a: number): void;
  getCourseId(): number;
  setCourseId(a: number): void;
  toObject(): RepositoryRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => RepositoryRequest;
}

export namespace RepositoryRequest {
  export type AsObject = {
    Id: number;
    Type: Repository.RepoType;
    DirectoryId: number;
    RepositoryId: number;
    UserId: number;
    CourseId: number;
  }
}

export class Submission {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getAssignmentId(): number;
  setAssignmentId(a: number): void;
  getUserId(): number;
  setUserId(a: number): void;
  getGroupId(): number;
  setGroupId(a: number): void;
  getScore(): number;
  setScore(a: number): void;
  getScoreObjects(): string;
  setScoreObjects(a: string): void;
  getBuildInfo(): string;
  setBuildInfo(a: string): void;
  getCommitHash(): string;
  setCommitHash(a: string): void;
  getApproved(): boolean;
  setApproved(a: boolean): void;
  toObject(): Submission.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Submission;
}

export namespace Submission {
  export type AsObject = {
    Id: number;
    AssignmentId: number;
    UserId: number;
    GroupId: number;
    Score: number;
    ScoreObjects: string;
    BuildInfo: string;
    CommitHash: string;
    Approved: boolean;
  }
}

export class Submissions {
  constructor ();
  getSubmissionsList(): Submission[];
  setSubmissionsList(a: Submission[]): void;
  toObject(): Submissions.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Submissions;
}

export namespace Submissions {
  export type AsObject = {
    SubmissionsList: Submission[];
  }
}

export class URLResponse {
  constructor ();
  getUrl(): string;
  setUrl(a: string): void;
  toObject(): URLResponse.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => URLResponse;
}

export namespace URLResponse {
  export type AsObject = {
    Url: string;
  }
}

export class User {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getIsAdmin(): boolean;
  setIsAdmin(a: boolean): void;
  getName(): string;
  setName(a: string): void;
  getStudentId(): string;
  setStudentId(a: string): void;
  getEmail(): string;
  setEmail(a: string): void;
  getAvatarUrl(): string;
  setAvatarUrl(a: string): void;
  getRemoteIdentitiesList(): RemoteIdentity[];
  setRemoteIdentitiesList(a: RemoteIdentity[]): void;
  getEnrollmentsList(): Enrollment[];
  setEnrollmentsList(a: Enrollment[]): void;
  toObject(): User.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => User;
}

export namespace User {
  export type AsObject = {
    Id: number;
    IsAdmin: boolean;
    Name: string;
    StudentId: string;
    Email: string;
    AvatarUrl: string;
    RemoteIdentitiesList: RemoteIdentity[];
    EnrollmentsList: Enrollment[];
  }
}

export class Users {
  constructor ();
  getUsersList(): User[];
  setUsersList(a: User[]): void;
  toObject(): Users.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Users;
}

export namespace Users {
  export type AsObject = {
    UsersList: User[];
  }
}

export class Void {
  constructor ();
  toObject(): Void.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Void;
}

export namespace Void {
  export type AsObject = {
  }
}

export class Timestamp {
  constructor ();
  getSeconds(): number;
  setSeconds(a: number): void;
  getNanos(): number;
  setNanos(a: number): void;
  toObject(): Timestamp.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Timestamp;
}

export namespace Timestamp {
  export type AsObject = {
    Seconds: number;
    Nanos: number;
  }
}

