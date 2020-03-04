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

// Contains a course, a student/group enrollment and a list
// of all assignments and the last submission for each assignment
export interface IAssignmentLink {
    course: Course;
    link: Enrollment;
    assignments: IStudentSubmission[];
}

// Contains an assignment, a latest submission,
// and a name of the submitter (user or group)
export interface IStudentSubmission {
    assignment: Assignment;
    latest?: ISubmission;
    authorName: string;
}

// Contains a user and the relation to a single course.
export interface IUserRelation {
    user: User;
    link: Enrollment;
}

// Browser only objects END

// Lab submission results
export interface IBuildInfo {
    buildid: number;
    builddate: Date;
    buildlog: string;
    execTime: number;
}

// A single test case object
export interface ITestCases {
    TestName: string;
    Score: number;
    MaxScore: number;
    Weight: number;
}

// A student/group submission
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
    executionTime: number;
    buildLog: string;
    testCases: ITestCases[];
    approved: boolean;
}
