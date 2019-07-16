import { Assignment, Course, Enrollment, User } from "../proto/ag_pb";

export interface IUser {
    id: number;
    name: string;
    email: string;
    avatarurl: string;
    studentid: string;
    isadmin: boolean;
}

// Browser only objects START

export interface IAssignmentLink {
    /**
     * The course to the group
     */
    course: Course;
    /**
     * The relation between the group and the course.
     * Is null if there is none
     */
    link: Enrollment;
    /**
     * A list of all assignments and the last submission if there
     * is a relation between the group and the course which is
     * student or teacher
     */
    assignments: IStudentSubmission[];
}

/**
 * An interface which contains an assignment and the latest submission
 * for a spessific user.
 */
export interface IStudentSubmission {
    assignment: Assignment;
    latest?: ISubmission;
}

/**
 * An interface which contains a user and the relation to a signe course.
 * Usually returned when a course is given.
 */
export interface IUserRelation {
    user: User;
    link: Enrollment;
}

// Browser only objects END

export interface IBuildInfo {
    buildid: number;
    builddate: Date;
    buildlog: string;
    execTime: number;
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
 * INewGroup represent data structure for a new group
 */
export interface INewGroup {
    name: string;
    userids: number[];
}
