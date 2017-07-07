export interface IUser {
    id: number;
    firstName: string;
    lastName: string;
    email: string;
    personId: number;
    isAdmin: boolean;
}

export function isCourse(value: any): value is ICourse {
    return value
        && typeof value.id === "number"
        && typeof value.name === "string"
        && typeof value.tag === "string";
}

// Browser only objects START

export interface IUserCourse {
    course: ICourse;
    link?: ICourseUserLink;
    assignments: IStudentSubmission[];
}

export interface IUserCourseCollection {
    user: IUser;
    courses: IUserCourse;
}

export interface IStudentSubmission {
    assignment: IAssignment;
    latest?: ILabInfo;
}

// Browser only objects END

export interface ICourse {
    id: number;
    name: string;
    tag: string;
    year: string;
}

export interface IAssignment {
    id: number;
    courseId: number;
    name: string;
    start: Date;
    deadline: Date;
    end: Date;
    assignmentGroupId?: number;
}

export interface IAssignmentGroup {
    id: number;
    required: number;
}

export enum CourseUserState {
    pending = 0,
    student = 1,
    rejected = 2,
    teacher = 3,
}

export interface ICourseUserLink {
    personId: number;
    courseId: number;
    state: CourseUserState;
}

export interface ITestCases {
    name: string;
    score: number;
    points: number;
    weight: number;
}

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

export interface ICoursesWithAssignments {
    course: ICourse;
    labs: IAssignment[];
}
