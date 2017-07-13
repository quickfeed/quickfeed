import { IMap, MapHelper } from "../map";
import {
    CourseUserState,
    IAssignment,
    ICourse,
    ICourseUserLink,
    ILabInfo,
    IOrganization,
    isCourse,
    IStudentSubmission,
    IUser,
    IUserCourse,
    IUserRelation,

} from "../models";

import { UserManager } from "../managers";

export interface ICourseProvider {
    getCourses(): Promise<ICourse[]>;
    getAssignments(courseId: number): Promise<IMap<IAssignment>>;
    // getCoursesStudent(): Promise<ICourseUserLink[]>;
    getCoursesFor(user: IUser, state?: CourseUserState): Promise<ICourse[]>;
    getUsersForCourse(course: ICourse, state?: CourseUserState): Promise<IUser[]>;

    addUserToCourse(user: IUser, course: ICourse): Promise<boolean>;
    changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean>;

    createNewCourse(courseData: ICourse): Promise<boolean>;
    getCourse(id: number): Promise<ICourse | null>;
    updateCourse(courseId: number, courseData: ICourse): Promise<boolean>;
    // deleteCourse(id: number): Promise<boolean>;

    getAllLabInfos(): Promise<IMap<ILabInfo>>;
    getDirectories(provider: string): Promise<IOrganization[]>;
}

export class CourseManager {
    private courseProvider: ICourseProvider;

    constructor(courseProvider: ICourseProvider) {
        this.courseProvider = courseProvider;
    }

    /**
     * Adds a user to a course
     * @param user The user to be added to a course
     * @param course The course the user should be added to
     * @returns True if succeeded and false otherwise
     */
    public async addUserToCourse(user: IUser, course: ICourse): Promise<boolean> {
        return this.courseProvider.addUserToCourse(user, course);
    }

    /**
     * Get a course from and id
     * @param id The id of the course
     */
    public async getCourse(id: number): Promise<ICourse | null> {
        // const a = (await this.courseProvider.getCourses())[id];
        // if (a) {
        //     return a;
        // }
        // return null;
        return await this.courseProvider.getCourse(id);
    }

    /**
     * Get all the courses available at the server
     */
    public async getCourses(): Promise<ICourse[]> {
        // return MapHelper.toArray(await this.courseProvider.getCourses());
        return await this.courseProvider.getCourses();
    }

    public async getCoursesWithState(user: IUser): Promise<IUserCourse[]> {
        const userCourses = await this.getCoursesFor(user);
        const allCourses = await this.getCourses();
        const newMap = allCourses.map<IUserCourse>((ele) => {
            const link: ICourseUserLink | undefined = userCourses.find((sub) => sub.id === ele.id) ? {
                courseId: ele.id, personId: user.id, state: CourseUserState.student,
            } : undefined;

            return {
                assignments: [],
                course: ele,
                link,
            };
        });
        return newMap;
    }

    /**
     * Get all courses related to a user
     * @param user The user to get courses to
     * @param state Optional. The state the relations should be in, all if not present
     */
    public async getCoursesFor(user: IUser, state?: CourseUserState): Promise<ICourse[]> {
        return this.courseProvider.getCoursesFor(user, state);
    }

    /**
     * Get all userlinks for a single course
     * @param course The course userlinks should be retrived from
     * @param state Optinal. The state of the relation, all if not present
     */
    /*public async getUserLinksForCourse(course: ICourse, state?: CourseUserState): Promise<ICourseUserLink[]> {
     const users: ICourseUserLink[] = [];
     for (const c of await this.courseProvider.getCoursesStudent()) {
     if (course.id === c.courseId && (state === undefined || c.state === CourseUserState.student)) {
     users.push(c);
     }
     }
     return users;
     }*/

    /**
     * Retrives one assignment from a single course
     * @param course The course the assignment is in
     * @param assignmentId The id to the assignment
     */
    public async getAssignment(course: ICourse, assignmentId: number): Promise<IAssignment | null> {
        const temp = await this.courseProvider.getAssignments(course.id);
        const assign = temp[assignmentId];
        if (assign) {
            return assign;
        }
        return null;
    }

    /**
     * Get all assignments in a single course
     * @param courseId The course id or ICourse to retrive assignments from
     */
    public async getAssignments(courseId: number | ICourse): Promise<IAssignment[]> {
        if (isCourse(courseId)) {
            courseId = courseId.id;
        }
        return MapHelper.toArray(await this.courseProvider.getAssignments(courseId));
    }

    /**
     * Change the userstate for a relation between a course and a user
     * @param link The link to change state of
     * @param state The new state of the relation
     */
    public async changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean> {
        return this.courseProvider.changeUserState(link, state);
    }

    /**
     * Creates a new course in the backend
     * @param courseData The course information to create a course from
     */
    public async createNewCourse(courseData: ICourse): Promise<boolean> {
        return this.courseProvider.createNewCourse(courseData);
    }

    /**
     * Updates a course with new information
     * @param courseData The new information for the course
     */
    public async updateCourse(courseId: number, courseData: ICourse): Promise<boolean> {
        return await this.courseProvider.updateCourse(courseId, courseData);
    }

    /**
     * Load an IUserCourse object for a single user and a single course
     * @param student The student the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getStudentCourse(student: IUser, course: ICourse): Promise<IUserCourse | null> {
        /*const link = (await this.courseProvider.getCoursesStudent())
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
         }*/
        // TODO: Implement linkstate somewhere else
        const returnTemp: IUserCourse = {
            link: { personId: student.id, courseId: course.id, state: CourseUserState.student },
            assignments: [],
            course,
        };
        await this.fillLinks(student, returnTemp);
        return returnTemp;
        // return null;
    }

    /**
     * Loads a single IStudentSubmission for a student and an assignment.
     * This will contains information about an assignment and the lates
     * sumbission information related to that assignment.
     * @param student The student the information should be retrived from
     * @param assignment The assignment the data should be loaded for
     */
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

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     * @param student The student to load the information for
     */
    public async getStudentCourses(student: IUser): Promise<IUserCourse[]> {
        const links: IUserCourse[] = [];
        const userCourses = await this.courseProvider.getCoursesFor(student);
        for (const course of userCourses) {
            links.push({
                assignments: [],
                course,
                link: { courseId: course.id, personId: student.id, state: CourseUserState.student },
            });
        }

        for (const link of links) {
            await this.fillLinks(student, link);
        }
        return links;
    }

    /**
     * Retrives all users related to a single course
     * @param course The course to retrive userinformation to
     * @param userMan Usermanager to be able to get user information
     * @param state Optinal. The state of the user to course relation
     */
    public async getUsersForCourse(course: ICourse,
                                   userMan: UserManager,
                                   state?: CourseUserState): Promise<IUserRelation[]> {

        return (await this.courseProvider.getUsersForCourse(course, state)).map<IUserRelation>((user) => {
            return {
                link: { courseId: course.id, personId: user.id, state: CourseUserState.student },
                user,
            };
        });
    }

    /**
     * Get all available directories or organisations for a single provider
     * @param provider The provider to load information from, for instance github og gitlab
     */
    public async  getDirectories(provider: string): Promise<IOrganization[]> {
        return await this.courseProvider.getDirectories(provider);
    }

    /**
     * Add IStudentSubmissions to an IUserCourse
     * @param student The student
     * @param studentCourse The student course
     */
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
