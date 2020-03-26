import {
    IStudentLabsForCourse,
    IStudentLab,
    ISubmission,
    IUserRelation,
} from "../models";

import { Assignment, Course, Enrollment, Group, Organization, Repository, Status, User, Void } from '../../proto/ag_pb';
import { ILogger } from "./LogManager";

export interface ICourseProvider {
    getCourses(): Promise<Course[]>;
    getAssignments(courseID: number): Promise<Assignment[]>;
    getCoursesFor(user: User, status: Enrollment.UserStatus[]): Promise<Enrollment[]>;
    getUsersForCourse(course: Course, noGroupMemebers?: boolean, status?: Enrollment.UserStatus[]):
        Promise<Enrollment[]>;

    addUserToCourse(course: Course, user: User): Promise<boolean>;
    changeUserStatus(enrollment: Enrollment, status: Enrollment.UserStatus): Promise<Status>;
    approveAll(courseID: number): Promise<boolean>;

    createNewCourse(course: Course): Promise<Course | Status>;
    getCourse(courseID: number): Promise<Course | null>;
    updateCourse(course: Course): Promise<Void | Status>;

    getCourseGroups(courseID: number): Promise<Group[]>;
    updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<Status>;
    createGroup(courseID: number, name: string, users: number[]): Promise<Group | Status>;
    getGroup(groupID: number): Promise<Group | null>;
    deleteGroup(courseID: number, groupID: number): Promise<Status>;
    getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null>;
    updateGroup(group: Group): Promise<Status>;

    getAllLabInfos(courseID: number, userID: number): Promise<ISubmission[]>;
    getAllGroupLabInfos(courseID: number, groupID: number): Promise<ISubmission[]>;
    getCourseLabs(courseID: number, groupLabs: boolean): Promise<IStudentLabsForCourse[]>;
    getAllUserEnrollments(userID: number): Promise<Enrollment[]>;
    getOrganization(orgName: string): Promise<Organization | Status >;
    getProviders(): Promise<string[]>;
    updateAssignments(courseID: number): Promise<boolean>;
    updateSubmission(courseID: number, submissionID: number, approve: boolean): Promise<boolean>;
    rebuildSubmission(assignmentID: number, submissionID: number): Promise<ISubmission | null>;
    getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>>;

    isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean>;
}

export class CourseManager {
    private courseProvider: ICourseProvider;

    constructor(courseProvider: ICourseProvider, logger: ILogger) {
        this.courseProvider = courseProvider;
    }

    public async addUserToCourse(course: Course, user: User): Promise<boolean> {
        return this.courseProvider.addUserToCourse(course, user);
    }

    public async getCourse(courseID: number): Promise<Course | null> {
        return this.courseProvider.getCourse(courseID);
    }

    /**
     * Get all available courses
     */
    public async getCourses(): Promise<Course[]> {
        return this.courseProvider.getCourses();
    }

    public async getCoursesWithUserStatus(user: User): Promise<IStudentLabsForCourse[]> {
        const userCourses = await this.courseProvider.getCoursesFor(user, []);
        const newMap: IStudentLabsForCourse[] = [];
        userCourses.forEach((ele) => {
            const crs = ele.getCourse();
            if (crs) {
                newMap.push({
                    labs: [],
                    course: crs,
                    enrollment: ele,
                });
            }
        });
        return newMap;
    }

    /**
     * Get all courses where user is enrolled into
     */
    public async getCoursesFor(user: User, status: Enrollment.UserStatus[]): Promise<Course[]> {
        const courses: Course[] = [];
        const enrolList = await this.courseProvider.getCoursesFor(user, status);
        enrolList.forEach((ele) => {
            const crs = ele.getCourse();
            if (crs) {
                courses.push(crs);
            }
        });
        return courses;
    }

    /**
     * Get all assignments for a single course
     */
    public async getAssignments(courseID: number): Promise<Assignment[]> {
        return this.courseProvider.getAssignments(courseID);
    }

    /**
     * Update status of a course enrollment
     */
    public async changeUserStatus(enrollment: Enrollment, status: Enrollment.UserStatus): Promise<Status> {
        return this.courseProvider.changeUserStatus(enrollment, status);
    }

    /**
     * Approve all pending enrollments for a course
     */
    public async approveAll(courseID: number): Promise<boolean> {
        return this.courseProvider.approveAll(courseID);
    }

    /**
     * Creates a new course in the backend
     */
    public async createNewCourse(course: Course): Promise<Course | Status> {
        return this.courseProvider.createNewCourse(course);
    }

    /**
     * Updates a course with new information
     */
    public async updateCourse(course: Course): Promise<Void | Status> {
        return this.courseProvider.updateCourse(course);
    }

    /**
     * Retrives all course enrollments with the latest
     * lab submissions for all individual course assignments
     */
    public async getCourseLabs(courseID: number, groupLabs: boolean): Promise<IStudentLabsForCourse[]> {
        return this.courseProvider.getCourseLabs(courseID, groupLabs);
    }

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     */
    public async getStudentCourses(student: User, status: Enrollment.UserStatus[]): Promise<IStudentLabsForCourse[]> {
        const links: IStudentLabsForCourse[] = [];
        const enrollments = await this.courseProvider.getCoursesFor(student, status);
        for (const enrol of enrollments) {
            const crs = enrol.getCourse();
            if (crs) {
                links.push({
                    labs: [],
                    course: crs,
                    enrollment: enrol,
                });
            }
        }
        for (const link of links) {
            await this.fillLinks(link, student);
        }
        return links;
    }

    /**
     * Fetches all enrollments for the user, courses and groups are preloaded.
     */
    public async getAllUserEnrollments(userID: number): Promise<Enrollment[]> {
        return this.courseProvider.getAllUserEnrollments(userID);
    }

    /**
     * Retrives all users related to a single course
     */
    public async getUsersForCourse(
        course: Course,
        noGroupMemebers?: boolean,
        status?: Enrollment.UserStatus[]): Promise<IUserRelation[]> {
        const userlinks: IUserRelation[] = [];
        const enrolls = await this.courseProvider.getUsersForCourse(course, noGroupMemebers, status);
        enrolls.forEach((ele) => {
            const usr = ele.getUser();
            if (usr) {
                ele.setCourseid(course.getId());
                userlinks.push({
                    enrollment: ele,
                    user: usr,
                });
            }

        });
        return userlinks;
    }

    public async createGroup(courseID: number, name: string, users: number[]): Promise<Group | Status> {
        return this.courseProvider.createGroup(courseID, name, users);
    }

    public async updateGroup(group: Group): Promise<Status> {
        return this.courseProvider.updateGroup(group);
    }

    /**
     * getCourseGroup returns all course groups
     */
    public async getCourseGroups(courseID: number): Promise<Group[]> {
        return this.courseProvider.getCourseGroups(courseID);
    }

    /**
     * Load an IAssignmentLink object for a single group and a single course
     */
    public async getGroupCourse(group: Group, course: Course): Promise<IStudentLabsForCourse | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const enrol = new Enrollment();
            enrol.setGroupid(group.getId());
            enrol.setCourseid(course.getId());
            enrol.setGroup(group);
            const groupCourse: IStudentLabsForCourse = {
                enrollment: enrol,
                labs: [],
                course,
            };
            await this.fillLinks(groupCourse, undefined, group);
            return groupCourse;
        }
        return null;
    }

    public async getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null> {
        return this.courseProvider.getGroupByUserAndCourse(courseID, userID);
    }

    public async updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<Status> {
        return this.courseProvider.updateGroupStatus(groupID, status);
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        return this.courseProvider.getGroup(groupID);
    }

    public async deleteGroup(courseID: number, groupID: number): Promise<Status> {
        return this.courseProvider.deleteGroup(courseID, groupID);
    }

    /**
     * updateAssignments updates the assignments on the backend database
     * for the given course. The assignment data is collected from the
     * assignment.yml files found in the course's tests repository; there
     * should be one assignment.yml file per lab assignment.
     */
    public async updateAssignments(courseID: number): Promise<boolean> {
        return this.courseProvider.updateAssignments(courseID);
    }

    /**
     * Get a github organization by name
     */
    public async getOrganization(orgName: string): Promise<Organization | Status> {
        return this.courseProvider.getOrganization(orgName);
    }

    /**
     * Get all enabled providers
     */
    public async getProviders(): Promise<string[]> {
        return this.courseProvider.getProviders();
    }

    public async getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>> {
        return this.courseProvider.getRepositories(courseID, types);
    }

    public async rebuildSubmission(assignmentID: number, submissionID: number): Promise<ISubmission | null> {
        return this.courseProvider.rebuildSubmission(assignmentID, submissionID);
    }

    public async updateSubmission(courseID: number, submissionID: number, approved: boolean): Promise<boolean> {
        return this.courseProvider.updateSubmission(courseID, submissionID, approved);
    }

    public async isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean> {
        return this.courseProvider.isEmptyRepo(courseID, userID, groupID);
    }

    /**
     * Generates a list with last submission build info
     * for every course assignment for all course students.
     * Used when displaying all course lab submissions for all
     * course students on TeacherPage.
     */
    public async fillLabLinks(
        course: Course,
        labLinks: IStudentLabsForCourse[],
        assignments?: Assignment[],
        ): Promise<IStudentLabsForCourse[]> {

        if (!assignments) {
            assignments = await this.getAssignments(course.getId());
        }
        for (const studentLabs of labLinks) {
            studentLabs.course = course;

            let studentName = "";
            if (studentLabs.labs.length > 0) {
                studentName = studentLabs.labs[0].authorName;
            }

            for (const lab of studentLabs.labs) {
                const suggestedAssignment = assignments.find((asm) => lab.assignment.getId() === asm.getId());
                if (suggestedAssignment) {
                    lab.assignment = suggestedAssignment;
                }
            }

            // fill up cells for all assignments, even with no submissions,
            // to display properly in the table with lab results
            for (const asm of assignments) {
                const exists = studentLabs.labs.find((ele) => asm.getId() === ele.assignment.getId());
                if (!exists) {
                    const voidSubmission: IStudentLab = {
                        assignment: asm,
                        authorName: studentName,
                    };
                    studentLabs.labs.push(voidSubmission);
                }
            }
        }
        return labLinks;
    }

    /**
     * Add lab submissions for every course lab
     * to the given student or group enrollment.
     * Used on StudentPage.
     */
    public async fillLinks(courseLabLink: IStudentLabsForCourse, student?: User, group?: Group,  assignments?: Assignment[]): Promise<void> {
        if (!(student || group)) {
            return;
        }

        if (!assignments || assignments.length < 1) {
            assignments = await this.getAssignments(courseLabLink.course.getId());
            if (assignments.length < 1) {
                return;
            }
        }
        let submissions : ISubmission[] = [];
        let labAuthorName = "";
        let wantGroupLinks = false;
        if (student) {
            submissions =
                await this.courseProvider.getAllLabInfos(courseLabLink.course.getId(), student.getId());
            labAuthorName = student.getName();
        } else if (group) {
            submissions =
                await this.courseProvider.getAllGroupLabInfos(courseLabLink.course.getId(), group.getId());
            labAuthorName = group.getName();
            wantGroupLinks = true;
        } else {
            return;
        }

        for (const a of assignments) {
            if (a.getIsgrouplab() === wantGroupLinks) {
                const lab = submissions.find((sub) => sub.assignmentid === a.getId());
                courseLabLink.labs.push({ assignment: a, submission: lab, authorName: labAuthorName});
            }
        }
    }
}
