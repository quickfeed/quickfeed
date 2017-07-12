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
    getCourses(): Promise<IMap<ICourse>>;
    getAssignments(courseId: number): Promise<IMap<IAssignment>>;
    getCoursesStudent(): Promise<ICourseUserLink[]>;
    getCoursesFor(user: IUser, state?: CourseUserState): Promise<ICourse[]>;

    addUserToCourse(user: IUser, course: ICourse): Promise<boolean>;
    changeUserState(link: ICourseUserLink, state: CourseUserState): Promise<boolean>;

    createNewCourse(courseData: ICourse): Promise<boolean>;
    updateCourse(courseData: ICourse): Promise<boolean>;
    deleteCourse(id: number): Promise<boolean>;

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
        const a = (await this.courseProvider.getCourses())[id];
        if (a) {
            return a;
        }
        return null;
    }

    /**
     * Get all the courses available at the server
     */
    public async getCourses(): Promise<ICourse[]> {
        return MapHelper.toArray(await this.courseProvider.getCourses());
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
    public async getUserLinksForCourse(course: ICourse, state?: CourseUserState): Promise<ICourseUserLink[]> {
        const users: ICourseUserLink[] = [];
        for (const c of await this.courseProvider.getCoursesStudent()) {
            if (course.id === c.courseId && (state === undefined || c.state === CourseUserState.student)) {
                users.push(c);
            }
        }
        return users;
    }

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
    public async updateCourse(courseData: ICourse): Promise<boolean> {
        return this.courseProvider.updateCourse(courseData);
    }

    /**
     * Delete a course
     * @param id The id of the course to delete
     */
    public async deleteCourse(id: number): Promise<boolean> {
        // TODO: Should it be possible to delete a course
        return this.courseProvider.deleteCourse(id);
    }

    /**
     * Load an IUserCourse object for a single user and a single course
     * @param student The student the information should be retrived from
     * @param course The course the data should be loaded for
     */
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

    /**
     * Retrives all users related to a single course
     * @param course The course to retrive userinformation to
     * @param userMan Usermanager to be able to get user information
     * @param state Optinal. The state of the user to course relation
     */
    public async getUsersForCourse(
        course: ICourse,
        userMan: UserManager,
        state?: CourseUserState): Promise<IUserRelation[]> {

        const courseStds: ICourseUserLink[] =
            await this.getUserLinksForCourse(course, state);
        const users = await userMan.getUsersAsMap(courseStds.map((e) => e.personId));
        return courseStds.map<IUserRelation>((link) => {
            const user = users[link.personId];
            if (!user) {
                // TODO: See if we should have an error here or not
                throw new Error("Link exist witout a user object");
            }
            return {
                link,
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
