interface IUser {
    id: number;
    firstName: string;
    lastName: string;
    email: string;
    personId: number;
}

function isCourse(value: any): value is ICourse {
    return value
        && typeof value.id === "number"
        && typeof value.name === "string"
        && typeof value.tag === "string";
}

interface ICourse {
    id: number;
    name: string;
    tag: string;
}

interface IAssignment {
    id: number;
    courseId: number;
    name: string;
    start: Date;
    deadline: Date;
    end: Date;
}

interface ICourseStudent {
    personId: number;
    courseId: number;
}

interface ITestCases{
    name: string;
    score: number;
    points: number;
    weight: number;
}
interface ILabInfo{
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
export {IUser, isCourse, ICourse, IAssignment, ICourseStudent, ITestCases, ILabInfo};
