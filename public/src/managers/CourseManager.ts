import {
    IAssignmentLink,
    IStudentSubmission,
    ISubmission,
    IUserRelation,
} from "../models";

import { Assignment, Course, Enrollment, Group, Organization, Repository, Status, User, Void } from "../../proto/ag_pb";
import { ILogger } from "./LogManager";

export interface ICourseProvider {
    getCourses(): Promise<Course[]>;
    getAssignments(courseID: number): Promise<Assignment[]>;
    getCoursesFor(user: User, state: Enrollment.UserStatus[]): Promise<Enrollment[]>;
    getUsersForCourse(course: Course, noGroupMemebers?: boolean, state?: Enrollment.UserStatus[]):
        Promise<Enrollment[]>;

    addUserToCourse(course: Course, user: User): Promise<boolean>;
    changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<Status>;
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
    getCourseLabs(courseID: number, groupLabs: boolean): Promise<IAssignmentLink[]>;
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

    public async getCoursesWithState(user: User): Promise<IAssignmentLink[]> {
        const userCourses = await this.courseProvider.getCoursesFor(user, []);
        const newMap: IAssignmentLink[] = [];
        userCourses.forEach((ele) => {
            const crs = ele.getCourse();
            if (crs) {
                newMap.push({
                    assignments: [],
                    course: crs,
                    link: ele,
                });
            }
        });
        return newMap;
    }

    /**
     * Get all courses where user is enrolled into
     */
    public async getCoursesFor(user: User, state: Enrollment.UserStatus[]): Promise<Course[]> {
        const courses: Course[] = [];
        const enrolList = await this.courseProvider.getCoursesFor(user, state);
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
    public async changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<Status> {
        const ans = await this.courseProvider.changeUserState(link, state);
        return ans;
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
    public async getCourseLabs(courseID: number, groupLabs: boolean): Promise<IAssignmentLink[]> {
        return this.courseProvider.getCourseLabs(courseID, groupLabs);
    }

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     */
    public async getStudentCourses(student: User, state: Enrollment.UserStatus[]): Promise<IAssignmentLink[]> {
        const links: IAssignmentLink[] = [];
        const enrollments = await this.courseProvider.getCoursesFor(student, state);
        for (const enrol of enrollments) {
            const crs = enrol.getCourse();
            if (crs) {
                links.push({
                    assignments: [],
                    course: crs,
                    link: enrol,
                });
            }
        }
        for (const link of links) {
            await this.fillLinks(student, link);
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
        state?: Enrollment.UserStatus[]): Promise<IUserRelation[]> {
        const userlinks: IUserRelation[] = [];
        const enrolls = await this.courseProvider.getUsersForCourse(course, noGroupMemebers, state);
        enrolls.forEach((ele) => {
            const usr = ele.getUser();
            if (usr) {
                ele.setCourseid(course.getId());
                userlinks.push({
                    link: ele,
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
    public async getGroupCourse(group: Group, course: Course): Promise<IAssignmentLink | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const enrol = new Enrollment();
            enrol.setGroupid(group.getId());
            enrol.setCourseid(course.getId());
            enrol.setGroup(group);
            const groupCourse: IAssignmentLink = {
                link: enrol,
                assignments: [],
                course,
            };
            await this.fillLinksGroup(group, groupCourse);
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
     * for every course assignment for all course students
     */
    public async fillLabLinks(
        course: Course,
        labLinks: IAssignmentLink[],
        assignments?: Assignment[],
        ): Promise<IAssignmentLink[]> {

        if (!assignments) {
            assignments = await this.getAssignments(course.getId());
        }
        for (const studentLabs of labLinks) {
            studentLabs.course = course;

            let studentName = "";
            if (studentLabs.assignments.length > 0) {
                studentName = studentLabs.assignments[0].authorName;
            }

            for (const lab of studentLabs.assignments) {
                const suggestedAssignment = assignments.find((asm) => lab.assignment.getId() === asm.getId());
                if (suggestedAssignment) {
                    lab.assignment = suggestedAssignment;
                }
            }

            // fill up cells for all assignments, even with no submissions,
            // to display properly in the table with lab results
            for (const asm of assignments) {
                const exists = studentLabs.assignments.find((ele) => asm.getId() === ele.assignment.getId());
                if (!exists) {
                    const voidSubmission: IStudentSubmission = {
                        assignment: asm,
                        authorName: studentName,
                    };
                    studentLabs.assignments.push(voidSubmission);
                }
            }
        }
        return labLinks;
    }

    /**
     * Add IStudentSubmissions to an IAssignmentLink
     */
    public async fillLinks(student: User, studentCourse: IAssignmentLink, assignments?: Assignment[]): Promise<void> {
        if (!studentCourse.link) {
            return;
        }
        if (!assignments) {
            assignments = await this.getAssignments(studentCourse.course.getId());
        }
        if (assignments.length > 0) {
            const submissions =
                await this.courseProvider.getAllLabInfos(studentCourse.course.getId(), student.getId());

            for (const a of assignments) {
                if (!a.getIsgrouplab()) {
                    const submission = submissions.find((sub) => sub.assignmentid === a.getId());
                    studentCourse.assignments.push({ assignment: a, latest: submission, authorName: student.getName()});
                }
            }
        }
    }

    /**
     * Add group submissions to an IAssignmentLink
     */
    public async fillLinksGroup(group: Group, groupCourse: IAssignmentLink, assignments?: Assignment[]):
        Promise<void> {
        if (!groupCourse.link) {
            return;
        }
        if (!assignments) {
            assignments = await this.getAssignments(groupCourse.course.getId());
        }
        if (assignments.length > 0) {
            const submissions =
                await this.courseProvider.getAllGroupLabInfos(groupCourse.course.getId(), group.getId());
            for (const a of assignments) {
                if (a.getIsgrouplab()) {
                    const submission = submissions.find((sub) => sub.assignmentid === a.getId());
                    groupCourse.assignments.push({ assignment: a, latest: submission, authorName: group.getName() });
                }
            }
        }
    }
}
