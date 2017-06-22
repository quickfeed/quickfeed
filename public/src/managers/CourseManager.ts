import { ICourse, IAssignment, isCourse } from "../models";

interface ICourseProvider {
    getCourses(): ICourse[];
    getAssignments(courseId: number): IAssignment[];
}

class CourseManager {
    courseProvider: ICourseProvider;
    constructor(courseProvider: ICourseProvider){
        this.courseProvider = courseProvider;
    }

    getCourses():ICourse[]{
        return this.courseProvider.getCourses();
    }

    getAssignments(courseId: number): IAssignment[];
    getAssignments(course: ICourse): IAssignment[];
    getAssignments(courseId: number | ICourse): IAssignment[] {
        if (isCourse(courseId)){
            courseId = courseId.id;
            console.log(courseId);
        }
        return this.courseProvider.getAssignments(courseId);
    }
    
}

export {ICourseProvider, CourseManager}