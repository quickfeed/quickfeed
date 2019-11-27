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
     * The current course
     */
    course: Course;
    /**
     * The relation between the group and the course.
     */
    link: Enrollment;
    /**
     * A list of all assignments and the last submission for each
     */
    assignments: IStudentSubmission[];
}

/**
 * An interface which contains an assignment, a latest submission,
 * and a name of the submitter (user or group)
 */
export interface IStudentSubmission {
    assignment: Assignment;
    latest?: ISubmission;
    authorName: string;
}

/**
 * An interface which contains a user and the relation to a single course.
 */
export interface IUserRelation {
    user: User;
    link: Enrollment;
}

// Browser only objects END

/**
 * Lab submission results
 */
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
    TestName: string;
    Score: number;
    MaxScore: number;
    Weight: number;
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
