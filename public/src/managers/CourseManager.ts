import { IMap, MapHelper } from "../map";
import {
    CourseUserState,
    IAssignment,
    ICourse,
    ICourseUserLink,
    ILabInfo,
    isCourse,
    IStudentSubmission,
    IUser,
    IUserCourse,

} from "../models";

import { UserManager } from "../managers";

interface ICourseProvider {
    getCourses(): Promise<IMap<ICourse>>;
    getAssignments(courseId: number): Promise<IMap<IAssignment>>;
    getCoursesStudent(): Promise<ICourseUserLink[]>;
    addUserToCourse(user: IUser, course: ICourse): Promise<boolean>;
    changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean>;
    createNewCourse(courseData: ICourse): Promise<boolean>;
    getAllLabInfos(): Promise<IMap<ILabInfo>>;
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

    public async getRelationsFor(user: IUser, state?: CourseUserState): Promise<ICourseUserLink[]> {
        const cLinks: ICourseUserLink[] = [];

        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId && (state === undefined || c.state === CourseUserState.student)) {
                cLinks.push(c);
            }
        }
        return cLinks;
    }

    public async getCoursesFor(user: IUser, state?: CourseUserState): Promise<ICourse[]> {
        const cLinks: ICourseUserLink[] = [];

        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId && (state === undefined || c.state === CourseUserState.student)) {
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

    public async getUserLinksForCourse(course: ICourse, state?: CourseUserState): Promise<ICourseUserLink[]> {
        const users: ICourseUserLink[] = [];
        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (course.id === c.courseId && (state === undefined || c.state === CourseUserState.student)) {
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

    public async changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean> {
        return this.courseProvider.changeUserState(link, state);
    }

    public async createNewCourse(courseData: ICourse): Promise<boolean> {
        return this.courseProvider.createNewCourse(courseData);
    }

    public async getStudentCourse(student: IUser, course: ICourse): Promise<IUserCourse | null> {
        const link = (await this.courseProvider.getCoursesStudent())
            .find((val) => val.courseId === course.id && val.personId === student.id);
        if (link) {
            const assignments = this.courseProvider.getAssignments(course.id);
            const returnTemp: IUserCourse = {
                link,
                course,
                assignments: [],
            };
            await this.fillLinks(student, returnTemp);
            return returnTemp;
        }
        return null;
    }

    public async getUserSubmittions(student: IUser, assignment: IAssignment): Promise<IStudentSubmission> {
        const temp = MapHelper.find(await this.courseProvider.getAllLabInfos(),
            (ele) => ele.studentId === student.id && ele.assignmentId === assignment.id);
        if (temp) {
            return {
                assignment,
                latest: temp,
            };
        }
        return {
            assignment,
            latest: undefined,
        };
    }

    public async getStudentCourses(student: IUser): Promise<IUserCourse[]> {
        const allLinks = await this.courseProvider.getCoursesStudent();
        const allCourses = this.courseProvider.getCourses();
        const links: IUserCourse[] = [];

        MapHelper.forEach(await allCourses, (course) => {
            const curLink = allLinks.find((link) =>
                link.courseId === course.id && link.personId === student.id);
            links.push({
                assignments: [],
                course,
                link: curLink,
            });
        });

        for (const link of links) {
            await this.fillLinks(student, link);
        }
        return links;
    }

    public async getUsersForCourse(course: ICourse, userMan: UserManager, state?: CourseUserState): Promise<IUser[]> {
        const courseStds: ICourseUserLink[] =
            await this.getUserLinksForCourse(course, state);
        return await userMan.getUsers(courseStds.map((e) => e.personId));
    }

    private async fillLinks(student: IUser, studentCourse: IUserCourse): Promise<void> {
        if (!studentCourse.link) {
            return;
        }
        const allSubmissions: IStudentSubmission[] = [];
        const assigns = await this.getAssignments(studentCourse.course.id);
        for (const assign of assigns) {
            const submission = await this.getUserSubmittions(student, assign);
            if (submission) {
                studentCourse.assignments.push(submission);
            }
        }
    }
}

export { ICourseProvider, CourseManager };
