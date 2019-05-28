export class ActionRequest {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getUserid(): number;
  setUserid(a: number): void;
  getGroupid(): number;
  setGroupid(a: number): void;
  getCourseid(): number;
  setCourseid(a: number): void;
  getStatus(): Enrollment.UserStatus;
  setStatus(a: Enrollment.UserStatus): void;
  getGroupstatus(): Group.GroupStatus;
  setGroupstatus(a: Group.GroupStatus): void;
  toObject(): ActionRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => ActionRequest;
}

export namespace ActionRequest {
  export type AsObject = {
    Id: number;
    Userid: number;
    Groupid: number;
    Courseid: number;
    Status: Enrollment.UserStatus;
    Groupstatus: Group.GroupStatus;
  }
}

export class Assignment {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getCourseid(): number;
  setCourseid(a: number): void;
  getName(): string;
  setName(a: string): void;
  getLanguage(): string;
  setLanguage(a: string): void;
  getDeadline(): Timestamp;
  setDeadline(a: Timestamp): void;
  getAutoapprove(): boolean;
  setAutoapprove(a: boolean): void;
  getOrder(): number;
  setOrder(a: number): void;
  getIsgrouplab(): boolean;
  setIsgrouplab(a: boolean): void;
  getSubmission(): Submission;
  setSubmission(a: Submission): void;
  toObject(): Assignment.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Assignment;
}

export namespace Assignment {
  export type AsObject = {
    Id: number;
    Courseid: number;
    Name: string;
    Language: string;
    Deadline: Timestamp;
    Autoapprove: boolean;
    Order: number;
    Isgrouplab: boolean;
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
  getCoursecreatorid(): number;
  setCoursecreatorid(a: number): void;
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
  getDirectoryid(): number;
  setDirectoryid(a: number): void;
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
    Coursecreatorid: number;
    Name: string;
    Code: string;
    Year: number;
    Tag: string;
    Provider: string;
    Directoryid: number;
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
  getCourseid(): number;
  setCourseid(a: number): void;
  toObject(): DirectoryRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => DirectoryRequest;
}

export namespace DirectoryRequest {
  export type AsObject = {
    Provider: string;
    Courseid: number;
  }
}

export class Enrollment {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getCourseid(): number;
  setCourseid(a: number): void;
  getUserid(): number;
  setUserid(a: number): void;
  getGroupid(): number;
  setGroupid(a: number): void;
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
    Courseid: number;
    Userid: number;
    Groupid: number;
    User: User;
    Course: Course;
    Group: Group;
    Status: Enrollment.UserStatus;
  }

  export enum UserStatus { 
    Pending = 0,
    Rejected = 1,
    Student = 2,
    Teacher = 3,
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
  getCourseid(): number;
  setCourseid(a: number): void;
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
    Courseid: number;
    Status: Group.GroupStatus;
    UsersList: User[];
    EnrollmentsList: Enrollment[];
  }

  export enum GroupStatus { 
    Pending = 0,
    Rejected = 1,
    Approved = 2,
    Deleted = 3,
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
  getGroupstatusesList(): Group.GroupStatus[];
  setGroupstatusesList(a: Group.GroupStatus[]): void;
  toObject(): RecordRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => RecordRequest;
}

export namespace RecordRequest {
  export type AsObject = {
    Id: number;
    StatusesList: Enrollment.UserStatus[];
    GroupstatusesList: Group.GroupStatus[];
  }
}

export class RemoteIdentity {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getProvider(): string;
  setProvider(a: string): void;
  getRemoteid(): number;
  setRemoteid(a: number): void;
  getAccesstoken(): string;
  setAccesstoken(a: string): void;
  getUserid(): number;
  setUserid(a: number): void;
  toObject(): RemoteIdentity.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => RemoteIdentity;
}

export namespace RemoteIdentity {
  export type AsObject = {
    Id: number;
    Provider: string;
    Remoteid: number;
    Accesstoken: string;
    Userid: number;
  }
}

export class Repository {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getDirectoryid(): number;
  setDirectoryid(a: number): void;
  getRepositoryid(): number;
  setRepositoryid(a: number): void;
  getUserid(): number;
  setUserid(a: number): void;
  getGroupid(): number;
  setGroupid(a: number): void;
  getHtmlurl(): string;
  setHtmlurl(a: string): void;
  getRepotype(): Repository.RepoType;
  setRepotype(a: Repository.RepoType): void;
  toObject(): Repository.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Repository;
}

export namespace Repository {
  export type AsObject = {
    Id: number;
    Directoryid: number;
    Repositoryid: number;
    Userid: number;
    Groupid: number;
    Htmlurl: string;
    Repotype: Repository.RepoType;
  }

  export enum RepoType { 
    User = 0,
    Assignment = 1,
    Tests = 2,
    Solution = 3,
    CourseInfo = 4,
  }
}

export class RepositoryRequest {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getType(): Repository.RepoType;
  setType(a: Repository.RepoType): void;
  getDirectoryid(): number;
  setDirectoryid(a: number): void;
  getRepositoryid(): number;
  setRepositoryid(a: number): void;
  getUserid(): number;
  setUserid(a: number): void;
  getCourseid(): number;
  setCourseid(a: number): void;
  toObject(): RepositoryRequest.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => RepositoryRequest;
}

export namespace RepositoryRequest {
  export type AsObject = {
    Id: number;
    Type: Repository.RepoType;
    Directoryid: number;
    Repositoryid: number;
    Userid: number;
    Courseid: number;
  }
}

export class Submission {
  constructor ();
  getId(): number;
  setId(a: number): void;
  getAssignmentid(): number;
  setAssignmentid(a: number): void;
  getUserid(): number;
  setUserid(a: number): void;
  getGroupid(): number;
  setGroupid(a: number): void;
  getScore(): number;
  setScore(a: number): void;
  getScoreobjects(): string;
  setScoreobjects(a: string): void;
  getBuildinfo(): string;
  setBuildinfo(a: string): void;
  getCommithash(): string;
  setCommithash(a: string): void;
  getApproved(): boolean;
  setApproved(a: boolean): void;
  toObject(): Submission.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => Submission;
}

export namespace Submission {
  export type AsObject = {
    Id: number;
    Assignmentid: number;
    Userid: number;
    Groupid: number;
    Score: number;
    Scoreobjects: string;
    Buildinfo: string;
    Commithash: string;
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
  getIsadmin(): boolean;
  setIsadmin(a: boolean): void;
  getName(): string;
  setName(a: string): void;
  getStudentid(): string;
  setStudentid(a: string): void;
  getEmail(): string;
  setEmail(a: string): void;
  getAvatarurl(): string;
  setAvatarurl(a: string): void;
  getRemoteidentitiesList(): RemoteIdentity[];
  setRemoteidentitiesList(a: RemoteIdentity[]): void;
  getEnrollmentsList(): Enrollment[];
  setEnrollmentsList(a: Enrollment[]): void;
  toObject(): User.AsObject;
  serializeBinary(): Uint8Array;
  static deserializeBinary: (bytes: {}) => User;
}

export namespace User {
  export type AsObject = {
    Id: number;
    Isadmin: boolean;
    Name: string;
    Studentid: string;
    Email: string;
    Avatarurl: string;
    RemoteidentitiesList: RemoteIdentity[];
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

