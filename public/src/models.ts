interface IUser {
    id: number;
    firstName: string;
    lastName: string;
    email: string;
    personId: number;
}

function isCourse(value: any): value is ICourse{
    return value
        && typeof value.id === "number"
        && typeof value.name === "string"
        && typeof value.tag === "string";
}

interface ICourse{
    id: number;
    name: string;
    tag: string;
}

interface IAssignment{
    id: number;
    courceId: number;
    name: string;
    start: Date;
    deadline: Date;
    end: Date;
}

interface ICourseStudent{
    personId: number;
    courseId: number;
}

export {IUser, isCourse, ICourse, IAssignment, ICourseStudent};
