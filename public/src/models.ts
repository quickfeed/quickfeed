export interface IUser {
    id: number;
    firstName: string;
    lastName: string;
    email: string;
    personId: number;
}

export function isCourse(value: any): value is ICourse {
    return value
        && typeof value.id === "number"
        && typeof value.name === "string"
        && typeof value.tag === "string";
}

export interface ICourse {
    id: number;
    name: string;
    tag: string;
}

export interface IAssignment {
    id: number;
    courseId: number;
    name: string;
    start: Date;
    deadline: Date;
    end: Date;
}

export enum CourseStudentState {
    pending = 0,
    accepted = 1,
    rejected = 2,
}

export interface ICourseStudent {
    personId: number;
    courseId: number;
    state: CourseStudentState;
}

export interface ITestCases {
    name: string;
    score: number;
    points: number;
    weight: number;
}

export interface ILabInfo {
    lab: string;
    course: string;
    score: number;
    weight: number;
    test_cases: ITestCases[];
    pass_tests: number;
    fail_tests: number;
    exec_time: number;
    build_time: Date;
    build_id: number;
}

export interface ICoursesWithAssignments {
    course: ICourse;
    labs: IAssignment[];
}
