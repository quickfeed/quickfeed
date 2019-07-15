import {
    ICourseLinkAssignment,
    IGroupCourse,
    INewGroup,
    ISubmission,
    IUserCourse,
    IUserRelation,
} from "../models";

import { Assignment, Course, Enrollment, Group, Organization, Status, User, Void } from "../../proto/ag_pb";
import { UserManager } from "../managers";
import { ILogger } from "./LogManager";

export interface ICourseProvider {
    getCourses(): Promise<Course[]>;
    getAssignments(courseID: number): Promise<Assignment[]>;
    getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<ICourseEnrollment[]>;
    getUsersForCourse(course: Course, noGroupMemebers?: boolean, state?: Enrollment.UserStatus[]):
        Promise<IUserEnrollment[]>;

    addUserToCourse(user: User, course: Course): Promise<boolean>;
    changeUserState(link: Enrollment, state: Enrollment.UserStatus): Promise<boolean>;

    createNewCourse(courseData: Course): Promise<Course | Status>;
    getCourse(id: number): Promise<Course | null>;
    updateCourse(courseID: number, courseData: Course): Promise<Void | Status>;

    getCourseGroups(courseID: number): Promise<Group[]>;
    updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<boolean>;
    createGroup(groupData: INewGroup, courseId: number): Promise<Group | Status>;
    getGroup(groupID: number): Promise<Group | null>;
    deleteGroup(groupID: number): Promise<boolean>;
    getGroupByUserAndCourse(userid: number, courseid: number): Promise<Group | null>;
    updateGroup(groupData: Group): Promise<Status>;

    getAllLabInfos(courseID: number, userId: number): Promise<ISubmission[]>;
    getAllGroupLabInfos(courseID: number, groupID: number): Promise<ISubmission[]>;
    getOrganizations(provider: string): Promise<Organization[]>;
    getProviders(): Promise<string[]>;
    updateAssignments(courseID: number): Promise<boolean>;
    approveSubmission(submissionID: number, courseID: number): Promise<void>;
    getRepositoryURL(cid: number, type: number): Promise<string>;
}

export function isUserEnrollment(enroll: IEnrollment): enroll is ICourseEnrollment {
    if ((enroll as any).course) {
        return true;
    }
    return false;
}

export function isCourseEnrollment(enroll: IEnrollment): enroll is IUserEnrollment {
    if ((enroll as any).user) {
        return true;
    }
    return false;
}

export interface ICourseEnrollment extends IEnrollment {
    course: Course;
}

export interface IUserEnrollment extends IEnrollment {
    user: User;
    status: Enrollment.UserStatus;
}

export interface IEnrollment {
    userid: number;
    courseid: number;
    status?: Enrollment.UserStatus;

    course?: Course;
    user?: User;
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
    public async addUserToCourse(user: User, course: Course): Promise<boolean> {
        return this.courseProvider.addUserToCourse(user, course);
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

    public async getCoursesWithState(user: User): Promise<IUserCourse[]> {
        const userCourses = await this.courseProvider.getCoursesFor(user);
        const newMap = userCourses.map<IUserCourse>((ele) => {
            const enrol = new Enrollment();
            enrol.setUserid(ele.userid);
            enrol.setCourseid(ele.courseid);
            if (ele.status) {
                enrol.setStatus(ele.status);
            }
            return {
                assignments: [],
                course: ele.course,
                link: ele.status !== undefined ? enrol : undefined,
            };
        });
        return newMap;
    }

    /**
     * Get all courses related to a user
     * @param user The user to get courses to
     * @param state Optional. The state the relations should be in, all if not present
     */
    public async getCoursesFor(user: User, state?: Enrollment.UserStatus[]): Promise<Course[]> {
        return (await this.courseProvider.getCoursesFor(user, state)).map((ele) => ele.course);
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
        return this.courseProvider.changeUserState(link, state);
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
     * @param courseData The new information for the course
     */
    public async updateCourse(courseID: number, courseData: Course): Promise<Void | Status> {
        return this.courseProvider.updateCourse(courseID, courseData);
    }

    /**
     * Load an IUserCourse object for a single user and a single course
     * @param student The student the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getStudentCourseForTeacher(student: IUserRelation, course: Course, assignments: Assignment[]):
        Promise<IUserCourse | null> {
        const enrol = new Enrollment();
        enrol.setUserid(student.user.getId());
        enrol.setCourseid(course.getId());
        enrol.setStatus(student.link.getStatus());

        const userCourse: IUserCourse = {
            link: enrol,
            assignments: [],
            course,
        };
        await this.fillLinks(student.user, userCourse, assignments);
        return userCourse;
    }

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     * @param student The student to load the information for
     */
    public async getStudentCourses(student: User, state?: Enrollment.UserStatus[]): Promise<ICourseLinkAssignment[]> {
        const links: IUserCourse[] = [];
        const userCourses = await this.courseProvider.getCoursesFor(student, state);
        for (const course of userCourses) {
            const enrol = new Enrollment();
            enrol.setUserid(student.getId());
            enrol.setCourseid(course.courseid);
            if (course.status) {
                enrol.setStatus(course.status);
            }
            links.push({
                assignments: [],
                course: course.course,
                link: course.status !== undefined ? enrol : undefined,
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
     * @param state Optional. The state of the user to course relation
     */
    public async getUsersForCourse(
        course: Course,
        noGroupMemebers?: boolean,
        state?: Enrollment.UserStatus[]): Promise<IUserRelation[]> {

        return (await this.courseProvider.getUsersForCourse(course, noGroupMemebers, state)).map<IUserRelation>(
            (user) => {
                const enrol = new Enrollment();
                enrol.setUserid(user.userid);
                enrol.setCourseid(course.getId());
                enrol.setStatus(user.status);
                return {
                    link: enrol,
                    user: user.user,
                };
            });
    }

    public async createGroup(groupData: INewGroup, courseID: number): Promise<Group | Status> {
        return this.courseProvider.createGroup(groupData, courseID);
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
     * Load an IGroupCourse object for a single group and a single course
     * @param group The group the information should be retrived from
     * @param course The course the data should be loaded for
     */
    public async getGroupCourse(group: Group, course: Course): Promise<IGroupCourse | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const enrol = new Enrollment();
            enrol.setGroupid(group.getId());
            enrol.setCourseid(course.getId());
            enrol.setGroup(group);
            const groupCourse: IGroupCourse = {
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
        Promise<IGroupCourse | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const enrol = new Enrollment();
            enrol.setGroupid(group.getId());
            enrol.setCourseid(course.getId());
            enrol.setGroup(group);
            const groupCourse: IGroupCourse = {
                link: enrol,
                assignments: [],
                course,
            };
            await this.fillLinksGroup(group, groupCourse, assignments);
            return groupCourse;
        }
        return null;
    }

    public async getGroupByUserAndCourse(userID: number, courseID: number): Promise<Group | null> {
        return this.courseProvider.getGroupByUserAndCourse(userID, courseID);
    }

    public async updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<boolean> {
        return this.courseProvider.updateGroupStatus(groupID, status);
    }

    public async getGroup(groupID: number): Promise<Group | null> {
        return this.courseProvider.getGroup(groupID);
    }

    public async deleteGroup(groupID: number): Promise<boolean> {
        return this.courseProvider.deleteGroup(groupID);
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

    /**
     * Get all available directories or organisations for a single provider
     * @param provider The provider to load information from, for instance github og gitlab
     */
    public async getOrganizations(provider: string): Promise<Organization[]> {
        return this.courseProvider.getOrganizations(provider);
    }

    public async getProviders(): Promise<string[]> {
        return this.courseProvider.getProviders();
    }

    public async getRepositoryURL(cid: number, type: number): Promise<string> {
        return this.courseProvider.getRepositoryURL(cid, type);
    }

    public async approveSubmission(submissionID: number, courseID: number): Promise<void> {
        return this.courseProvider.approveSubmission(submissionID, courseID);
    }

    /**
     * Add IStudentSubmissions to an IUserCourse
     * @param student The student
     * @param studentCourse The student course
     */
    private async fillLinks(student: User, studentCourse: IUserCourse, assignments?: Assignment[]): Promise<void> {
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
                studentCourse.assignments.push({ assignment: a, latest: submission });
            }
        }
    }

    /**
     * Add IStudentSubmissions to an IUserCourse
     * @param group The group
     * @param groupCourse The group course
     */
    private async fillLinksGroup(group: Group, groupCourse: IGroupCourse, assignments?: Assignment[]): Promise<void> {
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
                groupCourse.assignments.push({ assignment: a, latest: submission });
            }
        }
    }
}
