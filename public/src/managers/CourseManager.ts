import { ArrayHelper } from "../helper";
import { IAssignment, ICourse, ICourseStudent, isCourse, IUser } from "../models";

interface ICourseProvider {
    getCourses(): ICourse[];
    getAssignments(courseId: number): IAssignment[];
    getCoursesStudent(): ICourseStudent[];
    getCourseByTag(tag: string): ICourse | null;
}

class CourseManager {
    private courseProvider: ICourseProvider;
    constructor(courseProvider: ICourseProvider) {
        this.courseProvider = courseProvider;
    }

    public getCourse(id: number): ICourse | null {
        return ArrayHelper.find(this.getCourses(), (a) => a.id === id);
    }

    public getCourses(): ICourse[] {
        return this.courseProvider.getCourses();
    }

    // get a course by a given course tag
    getCourseByTag(tag: string): ICourse | null{
        return this.courseProvider.getCourseByTag(tag);
    }

    public getCoursesFor(user: IUser): ICourse[] {
        const cLinks: ICourseStudent[] = [];
        for (const c of this.courseProvider.getCoursesStudent()) {
            if (user.id === c.personId) {
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

}

export { ICourseProvider, CourseManager };
