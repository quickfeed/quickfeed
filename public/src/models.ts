import { Enrollment, Group, User } from "../proto/ag_pb";

// TODO(meling) can we delete these types, and instead use types from ag.proto?

export interface IUser {
    id: number;
    name: string;
    email: string;
    avatarurl: string;
    studentid: string;
    isadmin: boolean;
}

export interface IGroup {
    id: number;
    name: string;

}

/**
 * Checks if value is compatible with the ICourse interface
 * @param value A value to check if it is an ICourse
 */
export function isCourse(value: any): value is ICourse {
    return value
        && typeof value.id === "number"
        && typeof value.name === "string"
        && typeof value.tag === "string";
}

// Browser only objects START

export interface ICourseLinkAssignment {
    /**
     * The course to the group
     */
    course: ICourse;
    /**
     * The relation between the group and the course.
     * Is null if there is none
     */
    link?: ICourseUserLink | ICourseGroupLink;
    /**
     * A list of all assignments and the last submission if there
     * is a relation between the group and the course which is
     * student or teacher
     */
    assignments: IStudentSubmission[];
}

/**
 * An interface which contains both the course, the course
 * link, all assignments and the latest submission to a single user
 */
export interface IUserCourse extends ICourseLinkAssignment {
    link?: ICourseUserLink;
}

/**
 * An interface which contains both the course, the course
 * link, all assignments and the latest submission to a single group
 */
export interface IGroupCourse extends ICourseLinkAssignment {
    link?: ICourseGroupLink;
}

/**
 * Function to find out if object is a instance of IUserCourse.
 * Will return false if link is missing.
 * @param object
 */
export function isIUserCourse(test: IUserCourse | IGroupCourse | ICourseLinkAssignment): test is IUserCourse {
    const inter = test as IUserCourse;
    return inter.link ? (inter.link as ICourseUserLink).userid !== undefined : false;
}

/**
 * An IUserCourse instance which also contains the user it
 * is related to.
 * @see IUserCourse
 */
export interface IUserCourseWithUser {
    user: User;
    course: IUserCourse;
}

/**
 * An ICourseGroup instance which also contains the group it
 * is related to.
 * @see ICourseGroup
 */
export interface IGroupCourseWithGroup {
    group: Group;
    course: IGroupCourse;
}

/**
 * An interface which contains an assignment and the latest submission
 * for a spessific user.
 */
export interface IStudentSubmission {
    assignment: IAssignment;
    latest?: ISubmission;
}

/**
 * An interface which contains a user and the relation to a signe course.
 * Usualy returned when a course is given.
 */
export interface IUserRelation {
    user: User;
    link: ICourseUserLink;
}

/**
 * An interface which contains a user and the relation to a signe course.
 * Usualy returned when a course is given.
 */
export interface ICourseGroupRelation {
    group: Group;
    link: ICourseGroupLink;
}

// Browser only objects END

/**
 * Information about a single course
 */
export interface INewCourse {
    name: string;
    code: string;
    tag: string;
    year: number;
    provider: string;
    directoryid: number;
}

export interface ICourse extends INewCourse {
    id: number;
}

export interface IBuildInfo {
    buildid: number;
    builddate: Date;
    buildlog: string;
    execTime: number;
}

export interface ICourseWithEnrollStatus extends ICourse {
    enrolled: Enrollment.UserStatus;
}

/**
 * Information about a single assignment
 */
export interface IAssignment {
    id: number;
    courseid: number;
    name: string;
    language: string;
    deadline: Date;
    isgrouplab: boolean;

    // Not implemented yet
    // start: Date;
    // end: Date;

    assignmentGroupId?: number;
}

/**
 * Information about an assignment group
 * This is not implemented and is a feature for the future
 */
export interface IAssignmentGroup {
    id: number;
    required: number;
}

//TODO(vera): replacing these with protobuf enums
/**
 * A description of the relation between a user and a course
 */
/*
export enum CourseUserState {
    pending = 0,
    rejected = 1,
    student = 2,
    teacher = 3,
}
*/
/**
 * Status of a course group
 */
/*
export enum CourseGroupStatus {
    pending = 0,
    rejected = 1,
    approved = 2,
    deleted = 3,
}*/

/**
 * Status of a course group
 */
/*
export enum RepositoryType {
    UserRepo = 0,
    AssignmentsRepo = 1,
    TestsRepo = 2,
    SolutionsRepo = 3,
    CourseInfoRepo = 4,
}
*/
export function courseUserStateToString(state: Enrollment.UserStatus[]): string {
    return state.map((sta) => {
        switch (sta) {
            case Enrollment.UserStatus.PENDING:
                return "pending";
            case Enrollment.UserStatus.REJECTED:
                return "rejected";
            case Enrollment.UserStatus.STUDENT:
                return "student";
            case Enrollment.UserStatus.TEACHER:
                return "teacher";
            default:
                return "";
        }
    }).join(",");
}

export function courseGroupStateToString(state: Group.GroupStatus[]): string {
    return state.map((sta) => {
        switch (sta) {
            case Group.GroupStatus.PENDING:
                return "pending";
            case Group.GroupStatus.REJECTED:
                return "rejected";
            case Group.GroupStatus.APPROVED:
                return "approved";
            case Group.GroupStatus.DELETED:
                return "deleted";
            default:
                return "";
        }
    }).join(",");
}

/**
 * The relation description between a user and course
 */
export interface ICourseUserLink {
    userid: number;
    courseId: number;
    state: Enrollment.UserStatus;
}

/**
 * The relation description between a group and course
 */
export interface ICourseGroupLink {
    groupid: number;
    courseId: number;
    state: Group.GroupStatus;
}
/**
 * A description of a single test case object
 */
export interface ITestCases {
    name: string;
    score: number;
    points: number;
    weight: number;
}

/**
 * A description of a single user submission
 */
export interface ISubmission {
    id: number;
    userid: number;
    groupid: number;
    assignmentid: number;

    passedTests: number;
    failedTests: number;
    score: number;

    buildId: number;
    buildDate: Date;
    executetionTime: number;
    buildLog: string;
    testCases: ITestCases[];

    approved: boolean;

}

/**
 * A course description with all related assignments
 */
export interface ICoursesWithAssignments {
    course: ICourse;
    labs: IAssignment[];
}

/**
 * Description of an organization from the git web page
 */
/*
export interface IOrganization {
    id: number;
    path: string;
    avatar: string;
}
*/
/**
 * ICourseGroup represents a student group in a course
 */
/*
export interface ICourseGroup {
    id: number;
    name: string;
    status: Group.GroupStatus;
    courseid: number;
    users: User[];
}*/

/**
 * INewGroup represent data structure for a new group
 */
export interface INewGroup {
    name: string;
    userids: number[];
}

/**
 * IStatusCode represent the status code returns from serverside
 */
export interface IStatusCode {
    statusCode: number;
}

/**
 * IError represents server side error object
 */
export interface IError extends IStatusCode {
    data?: any;
}

/**
 * Checks if value is compatible with the IError interface
 * @param item A value to check if it is an IError
 */
export function isError(item: any): item is IError {
    return item && typeof item.statusCode === "number";
}
