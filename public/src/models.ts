import { Assignment, Course, Enrollment, User, GradingBenchmark } from '../proto/ag_pb';

export interface IUser {
    id: number;
    name: string;
    email: string;
    avatarurl: string;
    studentid: string;
    isadmin: boolean;
}

// Browser only objects START

// Contains a course, a student/group enrollment, and a list
// of all assignments and the last submission for each assignment
export interface IStudentLabsForCourse {
    course: Course;
    enrollment: Enrollment;
    labs: IStudentLab[];
}

// Contains an assignment, a latest submission,
// and a name of the submitter (user or group)
export interface IStudentLab {
    assignment: Assignment;
    submission?: ISubmission;
    authorName: string;
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

export interface IReview {
    id: number;
    submissionID: number;
    reviewerID: number;
    review: GradingBenchmark[];
    feedback: string;
    ready: boolean;

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
    reviews: IReview[];
    feedbackReady: boolean;
}
