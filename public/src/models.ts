export interface IUser {
    id: number;
    firstname: string;
    lastname: string;
    email: string;
    personid: number;
    isadmin: boolean;
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

/**
 * An interface which contains bouth the course, the course
 * link, all assignments and the latest submission to a single user
 */
export interface IUserCourse {
    /**
     * The course to the user
     */
    course: ICourse;
    /**
     * The relation between the user and the course.
     * Is null if there is none
     */
    link?: ICourseUserLink;
    /**
     * A list of all assignments and the last submission if there
     * is a relation between the user and the course which is
     * student or teacher
     */
    assignments: IStudentSubmission[];
}

/**
 * An IUserCourse instance which also contains the user it
 * is related to.
 * @see IUserCourse
 */
export interface IUserCourseWithUser {
    user: IUser;
    course: IUserCourse;
}

/**
 * An interface which contains an assignment and the latest submission
 * for a spessific user.
 */
export interface IStudentSubmission {
    assignment: IAssignment;
    latest?: ILabInfo;
}

/**
 * An interface which contains a user and the relation to a signe course.
 * Usualy returned when a course is given.
 */
export interface IUserRelation {
    user: IUser;
    link: ICourseUserLink;
}

// Browser only objects END

/**
 * Information about a single course
 */
export interface ICourse {
    id: number;
    name: string;
    code: string;
    tag: string;
    year: number;
    provider: string;
    directoryid: number;
}

/**
 * Information about a single assignment
 */
export interface IAssignment {
    id: number;
    courseid: number;
    name: string;

    deadline: Date;

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
 * A description of the relation between a user and a course
 */
export enum CourseUserState {
    pending = 0,
    rejected = 1,
    student = 2,
    teacher = 3,
}

export function courseUserStateToString(state: CourseUserState): string {
    switch (state) {
        case CourseUserState.pending:
            return "pending";
        case CourseUserState.rejected:
            return "rejected";
        case CourseUserState.student:
            return "accepted";
        case CourseUserState.teacher:
            return "accepted";
        default:
            return "";
    }
}

/**
 * The relation description between a user and course
 */
export interface ICourseUserLink {
    userid: number;
    courseId: number;
    state: CourseUserState;
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
export interface ILabInfo {
    id: number;
    studentId: number;
    assignmentId: number;

    passedTests: number;
    failedTests: number;
    score: number;
    weight: number;

    buildId: number;
    buildDate: Date;
    executetionTime: number;
    buildLog: string;
    testCases: ITestCases[];

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
export interface IOrganization {
    id: number;
    path: string;
    avatar: string;
}
