import { ICourse, IAssignment, isCourse, IUser, ICourseStudent } from "../models";

interface ICourseProvider {
    getCourses(): ICourse[];
    getAssignments(courseId: number): IAssignment[];
    getCoursesStudent(): ICourseStudent[];
    getCourseByTag(tag: string): ICourse | null;
}

class CourseManager {
    courseProvider: ICourseProvider;
    constructor(courseProvider: ICourseProvider){
        this.courseProvider = courseProvider;
    }

    getCourses():ICourse[]{
        return this.courseProvider.getCourses();
    }

    // get a course by a given course tag
    getCourseByTag(tag: string): ICourse | null{
        return this.courseProvider.getCourseByTag(tag);
    }

    getCoursesFor(user: IUser): ICourse[] {
        let cLinks: ICourseStudent[] = [];
        for(let c of this.courseProvider.getCoursesStudent()){
            if (user.id === c.personId){
                cLinks.push(c);
            }
        }
        let courses: ICourse[] = [];
        for(let c of this.getCourses()){
            for(let link of cLinks){
                if (c.id === link.courseId){
                    courses.push(c);
                    break;
                }
            }   
        }
        return courses;
    }

    getAssignment(course: ICourse, assignmentId: number): IAssignment | null{
        let temp = this.getAssignments(course);
        console.log(temp);
        for(let a of temp){
            if (a.id === assignmentId){
                return a;
            }
        }
        return null;
    }

    getAssignments(courseId: number): IAssignment[];
    getAssignments(course: ICourse): IAssignment[];
    getAssignments(courseId: number | ICourse): IAssignment[] {
        if (isCourse(courseId)){
            courseId = courseId.id;
        }
        return this.courseProvider.getAssignments(courseId);
    }
    
}

export {ICourseProvider, CourseManager}