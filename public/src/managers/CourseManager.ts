import { ArrayHelper } from "../helper";
import { CourseStudentState, IAssignment, ICourse, ICourseStudent, isCourse, IUser } from "../models";

interface ICourseProvider {
    getCourses(): ICourse[];
    getAssignments(courseId: number): IAssignment[];
    getCoursesStudent(): ICourseStudent[];
    getCourseByTag(tag: string): ICourse | null;
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
        return ArrayHelper.find(this.getCourses(), (a) => a.id === id);
    }

    public getCourses(): ICourse[] {
        return this.courseProvider.getCourses();
    }

    // get a course by a given course tag
    public getCourseByTag(tag: string): ICourse | null {
        return this.courseProvider.getCourseByTag(tag);
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
        for (const c of this.getCourses()) {
            for (const link of cLinks) {
                if (c.id === link.courseId) {
                    courses.push(c);
                    break;
                }
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
        for (const a of temp) {
            if (a.id === assignmentId) {
                return a;
            }
        }
        return null;
    }

    public getAssignments(courseId: number | ICourse): IAssignment[] {
        if (isCourse(courseId)) {
            courseId = courseId.id;
        }
        return this.courseProvider.getAssignments(courseId);
    }

    public changeUserState(link: ICourseStudent, state: CourseStudentState): void {
        this.courseProvider.changeUserState(link, state);
    }

}

export { ICourseProvider, CourseManager };
