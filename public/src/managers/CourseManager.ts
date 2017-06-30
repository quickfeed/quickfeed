import { ArrayHelper } from "../helper";
import { IMap, MapHelper } from "../map";
import { CourseStudentState, IAssignment, ICourse, ICourseStudent, isCourse, IUser } from "../models";

interface ICourseProvider {
    getCourses(): IMap<ICourse>;
    getAssignments(courseId: number): IMap<IAssignment>;
    getCoursesStudent(): ICourseStudent[];
    addUserToCourse(user: IUser, course: ICourse): void;
    changeUserState(link: ICourseStudent, state: CourseStudentState): void;
}

class CourseManager {
    private courseProvider: ICourseProvider;
    constructor(courseProvider: ICourseProvider) {
        this.courseProvider = courseProvider;
    }

    public addUserToCourse(user: IUser, course: ICourse): void {
        this.courseProvider.addUserToCourse(user, course);
    }

    public getCourse(id: number): ICourse | null {
        const a = this.getCourses()[id];
        if (a) {
            return a;
        }
        return null;
    }

    public getCourses(): ICourse[] {
        return MapHelper.toArray(this.courseProvider.getCourses());
    }

    public getRelationsFor(user: IUser, state?: CourseStudentState): ICourseStudent[] {
        const cLinks: ICourseStudent[] = [];

        for (const c of this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId && (state === undefined || c.state === CourseStudentState.accepted)) {
                cLinks.push(c);
            }
        }
        return cLinks;
    }

    public getCoursesFor(user: IUser, state?: CourseStudentState): ICourse[] {
        const cLinks: ICourseStudent[] = [];

        for (const c of this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId && (state === undefined || c.state === CourseStudentState.accepted)) {
                cLinks.push(c);
            }
        }
        const courses: ICourse[] = [];
        const tempCourses = this.getCourses();
        for (const link of cLinks) {
            const c = tempCourses[link.courseId];
            if (c) {
                courses.push(c);
            }
        }
        return courses;
    }

    public getUserIdsForCourse(course: ICourse): ICourseStudent[] {
        const users: ICourseStudent[] = [];
        for (const c of this.courseProvider.getCoursesStudent()) {
            if (course.id === c.courseId) {
                users.push(c);
            }
        }
        return users;
    }

    public getAssignment(course: ICourse, assignmentId: number): IAssignment | null {
        const temp = this.getAssignments(course);
        console.log(temp);
        if (temp[assignmentId]) {
            return temp[assignmentId];
        }
        return null;
    }

    public getAssignments(courseId: number | ICourse): IAssignment[] {
        if (isCourse(courseId)) {
            courseId = courseId.id;
        }
        return MapHelper.toArray(this.courseProvider.getAssignments(courseId));
    }

    public changeUserState(link: ICourseStudent, state: CourseStudentState): void {
        this.courseProvider.changeUserState(link, state);
    }

}

export { ICourseProvider, CourseManager };
