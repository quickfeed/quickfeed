import { Course, Enrollment, Group, User } from "../proto/ag_pb";

export interface IUser {
    id: number;
    name: string;
    email: string;
    avatarurl: string;
    studentid: string;
    isadmin: boolean;
}


/**
 * Checks if value is compatible with the ICourse interface
 * @param value A value to check if it is an ICourse
 */
export function isCourse(value: any): value is Course {
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
    course: Course;
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

export interface IBuildInfo {
    buildid: number;
    builddate: Date;
    buildlog: string;
    execTime: number;
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
    course: Course;
    labs: IAssignment[];
}

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
    message?: string;
}

/**
 * Checks if value is compatible with the IError interface
 * @param item A value to check if it is an IError
 */
export function isError(item: any): item is IError {
    return item && typeof item.statusCode === "number";
}
