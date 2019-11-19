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
    getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<Enrollment[]>;
    getUsersForCourse(course: Course, noGroupMemebers?: boolean, state?: Enrollment.UserStatus[]):
        Promise<Enrollment[]>;

    addUserToCourse(course: Course, user: User): Promise<boolean>;
    changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<boolean>;
    approveAll(courseID: number): Promise<boolean>;

    createNewCourse(courseData: Course): Promise<Course | Status>;
    getCourse(ID: number): Promise<Course | null>;
    updateCourse(course: Course): Promise<Void | Status>;

    getCourseGroups(courseID: number): Promise<Group[]>;
    updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<boolean>;
    createGroup(courseID: number, name: string, users: number[]): Promise<Group | Status>;
    getGroup(groupID: number): Promise<Group | null>;
    deleteGroup(courseID: number, groupID: number, withRepo: boolean): Promise<boolean>;
    getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null>;
    updateGroup(group: Group): Promise<Status>;

    getAllLabInfos(courseID: number, userID: number): Promise<ISubmission[]>;
    getAllGroupLabInfos(courseID: number, groupID: number): Promise<ISubmission[]>;
    getCourseLabs(courseID: number): Promise<IAssignmentLink[]>;

    getOrganization(orgName: string): Promise<Organization | Status >;
    getProviders(): Promise<string[]>;
    updateAssignments(courseID: number): Promise<boolean>;
    approveSubmission(courseID: number, submissionID: number): Promise<boolean>;
    refreshSubmission(ID: number): Promise<boolean>;
    getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>>;

    isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean>;
}

export class CourseManager {
    private courseProvider: ICourseProvider;

    constructor(courseProvider: ICourseProvider, logger: ILogger) {
        this.courseProvider = courseProvider;
    }

    /**
     * Adds a user to a course
     * @param user The user to be added to a course
     * @param course The course the user should be added to
     * @returns True if succeeded and false otherwise
     */
    public async addUserToCourse(course: Course, user: User): Promise<boolean> {
        return this.courseProvider.addUserToCourse(course, user);
    }

    /**
     * Get a course from and id
     * @param ID The id of the course
     */
    public async getCourse(ID: number): Promise<Course | null> {
        return this.courseProvider.getCourse(ID);
    }

    /**
     * Get all the courses available at the server
     */
    public async getCourses(): Promise<Course[]> {
        return this.courseProvider.getCourses();
    }

    public async getCoursesWithState(user: User): Promise<IAssignmentLink[]> {
        const userCourses = await this.courseProvider.getCoursesFor(user);
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
     * Get all courses related to a user
     * @param user The user to get courses to
     * @param state Optional. The state the relations should be in, all if not present
     */
    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<Course[]> {
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
     * Get all assignments in a single course
     * @param courseID The course id or ICourse to retrive assignments from
     */
    public async getAssignments(courseID: number): Promise<Assignment[]> {
        return this.courseProvider.getAssignments(courseID);
    }

    /**
     * Change the userstate for a relation between a course and a user
     * @param link The link to change state of
     * @param state The new state of the relation
     */
    public async changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<boolean> {
        const ans = await this.courseProvider.changeUserState(link, state);
        return ans;
    }

    public async approveAll(courseID: number): Promise<boolean> {
        return this.courseProvider.approveAll(courseID);
    }

    /**
     * Creates a new course in the backend
     * @param courseData The course information to create a course from
     */
    public async createNewCourse(courseData: Course): Promise<Course | Status> {
        return this.courseProvider.createNewCourse(courseData);
    }

    /**
     * Updates a course with new information
     * @param course The new information for the course
     */
    public async updateCourse(course: Course): Promise<Void | Status> {
        return this.courseProvider.updateCourse(course);
    }

    /**
     * Load an IAssignmentLink object for a single user and a single course
     * @param student The student the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getStudentCourseForTeacher(student: IUserRelation, course: Course, assignments: Assignment[]):
        Promise<IAssignmentLink | null> {
        const enrol = new Enrollment();
        enrol.setUserid(student.user.getId());
        enrol.setUser(student.user);
        enrol.setCourseid(course.getId());
        enrol.setCourse(course);
        enrol.setStatus(student.link.getStatus());

        const userCourse: IAssignmentLink = {
            link: enrol,
            assignments: [],
            course,
        };
        await this.fillLinks(student.user, userCourse, assignments);
        return userCourse;
    }

    public async getCourseLabs(courseID: number): Promise<IAssignmentLink[]> {
        return this.courseProvider.getCourseLabs(courseID);
    }

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     * @param student The student to load the information for
     */
    public async getStudentCourses(student: User, state?: Enrollment.UserStatus[]): Promise<IAssignmentLink[]> {
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
     * Retrives all users related to a single course
     * @param course The course to retrive userinformation to
     * @param state Optional. The state of the user to course relation
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

    public async updateGroup(groupData: Group): Promise<Status> {
        return this.courseProvider.updateGroup(groupData);
    }

    /**
     * getCourseGroup returns all the groups under a course
     * @param courseID course id of a course
     */
    public async getCourseGroups(courseID: number): Promise<Group[]> {
        return this.courseProvider.getCourseGroups(courseID);
    }

    /**
     * Load an IAssignmentLink object for a single group and a single course
     * @param group The group the information should be retrived from
     * @param course The course the data should be loaded for
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

    public async getGroupCourseForTeacher(group: Group, course: Course, assignments: Assignment[]):
        Promise<IAssignmentLink | null> {
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
            await this.fillLinksGroup(group, groupCourse, assignments);
            return groupCourse;
        }
        return null;
    }

    public async getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null> {
        return this.courseProvider.getGroupByUserAndCourse(courseID, userID);
    }

    public async updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<boolean> {
        return this.courseProvider.updateGroupStatus(groupID, status);
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        return this.courseProvider.getGroup(groupID);
    }

    public async deleteGroup(courseID: number, groupID: number, withRepo: boolean): Promise<boolean> {
        return this.courseProvider.deleteGroup(courseID, groupID, withRepo);
    }

    /**
     * updateAssignments updates the assignments on the backend database
     * for the given course. The assignment data is collected from the
     * assignment.yml files found in the course's tests repository; there
     * should be one assignment.yml file per lab assignment.
     * @param courseID course whose assignment to update
     */
    public async updateAssignments(courseID: number): Promise<boolean> {
        return this.courseProvider.updateAssignments(courseID);
    }

    public async getOrganization(orgName: string): Promise<Organization | Status> {
        return this.courseProvider.getOrganization(orgName);
    }

    /**
     * Get all available directories or organisations for a single provider
     * @param provider The provider to load information from, for instance github og gitlab
     */
    public async getProviders(): Promise<string[]> {
        return this.courseProvider.getProviders();
    }

    public async getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>> {
        return this.courseProvider.getRepositories(courseID, types);
    }

    public async refreshSubmission(ID: number): Promise<boolean> {
        return this.courseProvider.refreshSubmission(ID);
    }

    public async approveSubmission(courseID: number, submissionID: number): Promise<boolean> {
        return this.courseProvider.approveSubmission(courseID, submissionID);
    }

    public async isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean> {
        return this.courseProvider.isEmptyRepo(courseID, userID, groupID);
    }

    public async fillLabLinks(course: Course, labLinks: IAssignmentLink[], assignments?: Assignment[]): Promise<IAssignmentLink[]> {
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
     * @param student The student
     * @param studentCourse The student course
     */
    private async fillLinks(student: User, studentCourse: IAssignmentLink, assignments?: Assignment[]): Promise<void> {
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
                const submission = submissions.find((sub) => sub.assignmentid === a.getId());
                studentCourse.assignments.push({ assignment: a, latest: submission, authorName: student.getName() });
            }
        }
    }

    /**
     * Add IStudentSubmissions to an IAssignmentLink
     * @param group The group
     * @param groupCourse The group course
     */
    private async fillLinksGroup(group: Group, groupCourse: IAssignmentLink, assignments?: Assignment[]):
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
                const submission = submissions.find((sub) => sub.assignmentid === a.getId());
                groupCourse.assignments.push({ assignment: a, latest: submission, authorName: group.getName() });
            }
        }
    }
}
