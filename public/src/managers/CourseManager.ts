import { IMap, MapHelper } from "../map";
import { CourseStudentState, IAssignment, ICourse, ICourseStudent, isCourse, IUser } from "../models";

interface ICourseProvider {
    getCourses(): Promise<IMap<ICourse>>;
    getAssignments(courseId: number): Promise<IMap<IAssignment>>;
    getCoursesStudent(): Promise<ICourseStudent[]>;
    addUserToCourse(user: IUser, course: ICourse): Promise<boolean>;
    changeUserState(link: ICourseStudent, state: CourseStudentState): Promise<boolean>;
    createNewCourse(courseData: ICourse): Promise<boolean>;
}

class CourseManager {
    private courseProvider: ICourseProvider;

    constructor(courseProvider: ICourseProvider) {
        this.courseProvider = courseProvider;
    }

    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        return this.courseProvider.addUserToCourse(user, course);
    }

    public async getCourse(id: number): Promise<ICourse | null> {
        const a = (await this.getCourses())[id];
        if (a) {
            return a;
        }
        return null;
    }

    public async getCourses(): Promise<ICourse[]> {
        return MapHelper.toArray(await this.courseProvider.getCourses());
    }

    public async getRelationsFor(user: IUser, state?: CourseStudentState): Promise<ICourseStudent[]> {
        const cLinks: ICourseStudent[] = [];

        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId && (state === undefined || c.state === CourseStudentState.accepted)) {
                cLinks.push(c);
            }
        }
        return cLinks;
    }

    public async getCoursesFor(user: IUser, state?: CourseStudentState): Promise<ICourse[]> {
        const cLinks: ICourseStudent[] = [];

        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId && (state === undefined || c.state === CourseStudentState.accepted)) {
                cLinks.push(c);
            }
        }
        const courses: ICourse[] = [];
        const tempCourses = await this.getCourses();
        for (const link of cLinks) {
            const c = tempCourses[link.courseId];
            if (c) {
                courses.push(c);
            }
        }
        return courses;
    }

    public async getUserIdsForCourse(course: ICourse, state?: CourseStudentState): Promise<ICourseStudent[]> {
        const users: ICourseStudent[] = [];
        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (course.id === c.courseId && (state === undefined || c.state === CourseStudentState.accepted)) {
                users.push(c);
            }
        }
        return users;
    }

    public async getAssignment(course: ICourse, assignmentId: number): Promise<IAssignment | null> {
        const temp = await this.courseProvider.getAssignments(course.id);
        if (temp[assignmentId]) {
            return temp[assignmentId];
        }
        return null;
    }

    public async getAssignments(courseId: number | ICourse): Promise<IAssignment[]> {
        if (isCourse(courseId)) {
            courseId = courseId.id;
        }
        return MapHelper.toArray(await this.courseProvider.getAssignments(courseId));
    }

    public async changeUserState(link: ICourseStudent, state: CourseStudentState): Promise<boolean> {
        return this.courseProvider.changeUserState(link, state);
    }

    public async createNewCourse(courseData: ICourse): Promise<boolean> {
        return this.courseProvider.createNewCourse(courseData);
    }

}

export { ICourseProvider, CourseManager };
