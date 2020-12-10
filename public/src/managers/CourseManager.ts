import {
    IAllSubmissionsForEnrollment,
    ISubmissionLink,
    ISubmission,
} from "../models";

import {
    Assignment,
    Benchmarks,
    Course,
    Enrollment,
    Group,
    Organization,
    Repository,
    Status,
    User,
    Review,
    GradingBenchmark,
    GradingCriterion,
    SubmissionsForCourseRequest
} from "../../proto/ag_pb";
import { ILogger } from "./LogManager";
import { sortAssignmentsByOrder } from "../componentHelper";

export interface ICourseProvider {
    getCourses(): Promise<Course[]>;
    getAssignments(courseID: number): Promise<Assignment[]>;
    getCoursesForUser(user: User, status: Enrollment.UserStatus[]): Promise<Course[]>;
    getUsersForCourse(course: Course, withoutGroupMemebers?: boolean, withActivity?: boolean, status?: Enrollment.UserStatus[]):
        Promise<Enrollment[]>;

    addUserToCourse(course: Course, user: User): Promise<boolean>;
    changeUserStatus(enrollment: Enrollment, status: Enrollment.UserStatus): Promise<Status>;
    approveAll(courseID: number): Promise<boolean>;

    createNewCourse(course: Course): Promise<Course | Status>;
    getCourse(courseID: number): Promise<Course | null>;
    updateCourse(course: Course): Promise<Status>;
    updateCourseVisibility(enrol: Enrollment): Promise<boolean>;

    getGroupsForCourse(courseID: number): Promise<Group[]>;
    updateGroupStatus(groupID: number, status: Group.GroupStatus): Promise<Status>;
    createGroup(courseID: number, name: string, users: number[]): Promise<Group | Status>;
    getGroup(groupID: number): Promise<Group | null>;
    deleteGroup(courseID: number, groupID: number): Promise<Status>;
    getGroupByUserAndCourse(courseID: number, userID: number): Promise<Group | null>;
    updateGroup(group: Group): Promise<Status>;

    getSubmissionsByUser(courseID: number, userID: number): Promise<ISubmission[]>;
    getSubmissionsByGroup(courseID: number, groupID: number): Promise<ISubmission[]>;
    getSubmissionsByCourse(courseID: number, type: SubmissionsForCourseRequest.Type): Promise<IAllSubmissionsForEnrollment[]>;
    getEnrollmentsForUser(userID: number, statuses?: Enrollment.UserStatus[]): Promise<Enrollment[]>;
    getOrganization(orgName: string): Promise<Organization | Status >;
    getProviders(): Promise<string[]>;
    updateAssignments(courseID: number): Promise<boolean>;
    updateSubmission(courseID: number, submission: ISubmission): Promise<boolean>;
    updateSubmissions(assignmentID: number, courseID: number, score: number, release: boolean, approve: boolean): Promise<boolean>;
    rebuildSubmission(assignmentID: number, submissionID: number): Promise<ISubmission | null>;
    getRepositories(courseID: number, types: Repository.Type[]): Promise<Map<Repository.Type, string>>;
    isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean>;
    addNewBenchmark(bm: GradingBenchmark): Promise<GradingBenchmark | null>;
    addNewCriterion(c: GradingCriterion): Promise<GradingCriterion | null>;
    updateBenchmark(bm: GradingBenchmark): Promise<boolean>;
    updateCriterion(c: GradingCriterion): Promise<boolean>;
    deleteBenchmark(bm: GradingBenchmark): Promise<boolean>;
    deleteCriterion(c: GradingCriterion): Promise<boolean>;
    loadCriteria(assignmentID: number, courseID: number): Promise<GradingBenchmark[]>;
    addReview(r: Review, courseID: number): Promise<Review | null>;
    editReview(r: Review, courseID: number): Promise<boolean>;
    getReviewers(submissionID: number, courseID: number): Promise<User[]>
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

    public async getAllCoursesForEnrollmentPage(user: User): Promise<IAllSubmissionsForEnrollment[]> {
        const userCourses = await this.courseProvider.getEnrollmentsForUser(user.getId(), []);	        const userEnrollments = await this.courseProvider.getEnrollmentsForUser(user.getId(), []);
        const allCourses = await this.courseProvider.getCourses();
        const newMap: IAllSubmissionsForEnrollment[] = [];
        allCourses.forEach((crs) => {
            let enrol = userEnrollments.find(item => item.getCourseid() === crs.getId());
            if (!enrol) {
                enrol = new Enrollment();
                enrol.setCourseid(crs.getId());
                enrol.setUserid(user.getId());
                enrol.setUser(user);
                enrol.setCourse(crs);
                enrol.setStatus(Enrollment.UserStatus.NONE);
            }
            newMap.push({
                labs: [],
                course: crs,
                enrollment: enrol,
            });
        });
        return newMap;
    }
    /**
     * Get all courses where user is enrolled into
     */
    public async getCoursesForUser(user: User, status: Enrollment.UserStatus[], states: Enrollment.DisplayState[]): Promise<Course[]> {
        return this.courseProvider.getCoursesForUser(user, status);
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
    public async updateCourse(course: Course): Promise<Status> {
        return this.courseProvider.updateCourse(course);
    }

    public async updateCourseVisibility(enrol: Enrollment): Promise<boolean> {
        return this.courseProvider.updateCourseVisibility(enrol);
    }

    /**
     * Retrives all course enrollments with the latest
     * lab submissions for all individual course assignments
     */
    public async getSubmissionsByCourse(courseID: number, type: SubmissionsForCourseRequest.Type): Promise<IAllSubmissionsForEnrollment[]> {
        return this.courseProvider.getSubmissionsByCourse(courseID, type);
    }

    /**
     * Retrives all course relations, and courses related to a
     * a single student
     */
    public async getStudentCourses(student: User, status: Enrollment.UserStatus[]): Promise<IAllSubmissionsForEnrollment[]> {
        const links: IAllSubmissionsForEnrollment[] = [];
        const enrollments = await this.courseProvider.getEnrollmentsForUser(student.getId(), status);
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
    public async getEnrollmentsForUser(userID: number, statuses?: Enrollment.UserStatus[]): Promise<Enrollment[]> {
        return this.courseProvider.getEnrollmentsForUser(userID, statuses);
    }

    /**
     * Retrives all users related to a single course
     */
    public async getUsersForCourse(
        course: Course,
        withoutGroupMemebers?: boolean,
        withActivity?: boolean,
        status?: Enrollment.UserStatus[]): Promise<Enrollment[]> {
        return this.courseProvider.getUsersForCourse(course, withoutGroupMemebers, withActivity, status);
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
    public async getGroupsForCourse(courseID: number): Promise<Group[]> {
        return this.courseProvider.getGroupsForCourse(courseID);
    }

    /**
     * Load an IAssignmentLink object for a single group and a single course
     */
    public async getGroupCourse(group: Group, course: Course): Promise<IAllSubmissionsForEnrollment | null> {
        // Fetching group enrollment status
        if (group.getCourseid() === course.getId()) {
            const enrol = new Enrollment();
            enrol.setGroupid(group.getId());
            enrol.setCourseid(course.getId());
            enrol.setGroup(group);
            const groupCourse: IAllSubmissionsForEnrollment = {
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
     * assignment.yml files found in the course"s tests repository; there
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

    public async updateSubmission(courseID: number, submission: ISubmission): Promise<boolean> {
        return this.courseProvider.updateSubmission(courseID, submission);
    }

    public async isEmptyRepo(courseID: number, userID: number, groupID: number): Promise<boolean> {
        return this.courseProvider.isEmptyRepo(courseID, userID, groupID);
    };

    public async addNewBenchmark(bm: GradingBenchmark): Promise<GradingBenchmark | null> {
        return this.courseProvider.addNewBenchmark(bm);
    }

    public async addNewCriterion(c: GradingCriterion): Promise<GradingCriterion | null> {
        return this.courseProvider.addNewCriterion(c);
    }

    public async updateBenchmark(bm: GradingBenchmark): Promise<boolean> {
        return this.courseProvider.updateBenchmark(bm);
    }

    public async updateCriterion(c: GradingCriterion): Promise<boolean> {
        return this.courseProvider.updateCriterion(c);
    }

    public async deleteBenchmark(bm: GradingBenchmark): Promise<boolean> {
        return this.courseProvider.deleteBenchmark(bm);
    }
    public async deleteCriterion(c: GradingCriterion): Promise<boolean> {
        return this.courseProvider.deleteCriterion(c);
    }

    public async loadCriteria(assignmentID: number, courseID: number): Promise<GradingBenchmark[]> {
        return this.courseProvider.loadCriteria(assignmentID, courseID);
    }

    public async addReview(r: Review, courseID: number): Promise<Review | null> {
        return this.courseProvider.addReview(r, courseID);
    }

    public async editReview(r: Review, courseID: number): Promise<boolean> {
        return this.courseProvider.editReview(r, courseID);
    }

    public async getReviewers(submissionID: number, courseID: number): Promise<User[]> {
        return this.courseProvider.getReviewers(submissionID, courseID);
    }

    public async updateSubmissions(assignmentID: number, courseID: number, score: number, release: boolean, approve: boolean): Promise<boolean> {
        return this.courseProvider.updateSubmissions(assignmentID, courseID, score, release, approve);
    }

    /**
     * Generates a list with last submission build info
     * for every course assignment for all course students.
     * Used when displaying all course lab submissions for all
     * course students on TeacherPage.
     */
    public async fillLabLinks(
        course: Course,
        labLinks: IAllSubmissionsForEnrollment[],
        assignments?: Assignment[],
        ): Promise<IAllSubmissionsForEnrollment[]> {

        if (!assignments) {
            assignments = await this.getAssignments(course.getId());
        }
        assignments = sortAssignmentsByOrder(assignments);
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
                    const voidSubmission: ISubmissionLink = {
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
    public async fillLinks(courseLabLink: IAllSubmissionsForEnrollment, student?: User, group?: Group,  assignments?: Assignment[]): Promise<void> {
        if (!(student || group)) {
            return;
        }

        if (!assignments || assignments.length < 1) {
            assignments = await this.getAssignments(courseLabLink.course.getId());
            if (assignments.length < 1) {
                return;
            }
        }
        assignments = sortAssignmentsByOrder(assignments);
        let submissions : ISubmission[] = [];
        let labAuthorName = "";
        let wantGroupLinks = false;
        if (student) {
            submissions =
                await this.courseProvider.getSubmissionsByUser(courseLabLink.course.getId(), student.getId());
            labAuthorName = student.getName();
        } else if (group) {
            submissions =
                await this.courseProvider.getSubmissionsByGroup(courseLabLink.course.getId(), group.getId());
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
