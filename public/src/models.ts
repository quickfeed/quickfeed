import { Assignment, Course, Enrollment, Group, User } from "../proto/ag_pb";

export interface IUser {
    id: number;
    name: string;
    email: string;
    avatarurl: string;
    studentid: string;
    isadmin: boolean;
}

// Browser only objects START

export interface ICourseLinkAssignment {
    /**
     * The course to the group
     */
    course: Course;
    /**
     * The relation between the group and the course.
     * Is null if there is none
     */
    link?: Enrollment;
    /**
     * A list of all assignments and the last submission if there
     * is a relation between the group and the course which is
     * student or teacher
     */
    assignments: IStudentSubmission[];
}

/**
 * An interface which contains both the course, the course
 * link, all assignments and the latest submission to a single user
 */
export interface IUserCourse extends ICourseLinkAssignment {
    link?: Enrollment;
}

/**
 * An interface which contains both the course, the course
 * link, all assignments and the latest submission to a single group
 */
export interface IGroupCourse extends ICourseLinkAssignment {
    link?: Enrollment;
}

/**
 * An IUserCourse instance which also contains the user it
 * is related to.
 * @see IUserCourse
 */
export interface IUserCourseWithUser {
    user: User;
    course: IUserCourse;
}

/**
 * An ICourseGroup instance which also contains the group it
 * is related to.
 * @see ICourseGroup
 */
export interface IGroupCourseWithGroup {
    group: Group;
    course: IGroupCourse;
}

/**
 * An interface which contains an assignment and the latest submission
 * for a spessific user.
 */
export interface IStudentSubmission {
    assignment: Assignment;
    latest?: ISubmission;
}

/**
 * An interface which contains a user and the relation to a signe course.
 * Usually returned when a course is given.
 */
export interface IUserRelation {
    user: User;
    link: Enrollment;
}

// Browser only objects END

export interface IBuildInfo {
    buildid: number;
    builddate: Date;
    buildlog: string;
    execTime: number;
}

/**
 * A description of a single test case object
 */
export interface ITestCases {
    name: string;
    score: number;
    points: number;
    weight: number;
}

/**
 * A description of a single user submission
 */
export interface ISubmission {
    id: number;
    userid: number;
    groupid: number;
    assignmentid: number;

    passedTests: number;
    failedTests: number;
    score: number;

    buildId: number;
    buildDate: Date;
    executetionTime: number;
    buildLog: string;
    testCases: ITestCases[];

    approved: boolean;
}

/**
 * INewGroup represent data structure for a new group
 */
export interface INewGroup {
    name: string;
    userids: number[];
}
