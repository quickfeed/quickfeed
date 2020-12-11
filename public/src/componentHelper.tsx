import * as React from "react";
import { Assignment, Course, Enrollment, Group, Review, User, Submission, GradingBenchmark, GradingCriterion } from '../proto/ag_pb';
import { IAllSubmissionsForEnrollment, ISubmissionLink, ISubmission } from './models';

export function sortEnrollmentsByVisibility(enrols: Enrollment[], withHidden: boolean): Enrollment[] {
    let sorted: Enrollment[] = [];
    const active: Enrollment[] = [];
    const archived: Enrollment[] = [];
    enrols.forEach((enrol) => {
        switch (enrol.getState()) {
            case Enrollment.DisplayState.FAVORITE:
                sorted.push(enrol);
                break;
            case Enrollment.DisplayState.VISIBLE:
                active.push(enrol);
                break;
            case Enrollment.DisplayState.HIDDEN:
                if (withHidden) {
                    archived.push(enrol);
                }
                break;
            case Enrollment.DisplayState.UNSET:
                active.push(enrol);
                break;
        }
    })
    sorted = sorted.concat(active, archived);
    return sorted;
}

export function sortEnrollmentsByActivity(enrols: Enrollment[]): Enrollment[] {
    const teachers: Enrollment[] = [];
    const active: Enrollment[] = [];
    const inactive: Enrollment[] = [];
    enrols.forEach((enrol) => {
        if (enrol.getStatus() === Enrollment.UserStatus.TEACHER) {
            teachers.push(enrol);
        } else if (enrol.getLastactivitydate() === "") {
            inactive.push(enrol);
        } else {
            active.push(enrol);
        }
    });
    return teachers.concat(active, inactive);
}

export function sortStudentsForRelease<T>(fullList: T[], allSubmissions: Map<T, ISubmissionLink>, reviewers: number): T[] {
    const withReviews: T[] = [];
    const withSubmission: T[] = [];
    const noSubmissions: T[] = [];
    fullList.forEach(item => {
        const v = allSubmissions.get(item);
        if (v && v.submission && hasAllReviews(v.submission, reviewers)) {
            withReviews.push(item);
        } else if (v && v.submission) {
            withSubmission.push(item);
        } else {
            noSubmissions.push(item);
        }
    });
    return withReviews.concat(withSubmission, noSubmissions);
}

export function selectFromSubmissionLinks(allCourseLinks: IAllSubmissionsForEnrollment[], groupAssignment: boolean): (User | Group)[] {
    const list: (User | Group)[] = [];
    allCourseLinks.forEach(link => {
        const grp = link.enrollment.getGroup();
        const usr = link.enrollment.getUser();
        if (groupAssignment && grp) {
            list.push(grp);
        } else if (usr) {
            list.push(usr);
        }
    });
    return list;
}

// used in menus: ignores hidden courses
export function sortCoursesByVisibility(enrols: Enrollment[]): Course[] {
    let favorite: Course[] = [];
    const active: Course[] = [];
    enrols.forEach((e) => {
        const crs = e.getCourse();
        switch (e.getState()) {
            case Enrollment.DisplayState.FAVORITE:
                if (crs) favorite.push(crs);
                break;
            case Enrollment.DisplayState.VISIBLE:
                if (crs) active.push(crs);
                break;
            case Enrollment.DisplayState.UNSET:
                if (crs) active.push(crs);
            default:
                break;
        }
    });
    favorite = favorite.concat(active);
    return favorite;
}

export function sortAssignmentsByOrder(assignments: Assignment[]): Assignment[] {
    return assignments.sort((a, b) => a.getOrder() - b.getOrder());
}

export function sortUsersByAdminStatus(users: Enrollment[]): Enrollment[] {
    return users.sort((x, y) => ((x.getUser()?.getIsadmin() ?? false) < (y.getUser()?.getIsadmin() ?? false) ? 1 : -1));
}

export function getSlipDays(allLabs: IAllSubmissionsForEnrollment[], selected: ISubmission, forGroups: boolean): number {
    let days = 0;
    const wantID = forGroups ? selected.groupid : selected.userid;
    allLabs.forEach(item => {
        const haveID = forGroups ? item.enrollment.getGroupid() : item.enrollment.getUserid();
        if (haveID === wantID) {
            days = item.enrollment.getSlipdaysremaining();
        }
    });
    return days;
}

export function searchForStudents(enrols: Enrollment[], query: string): Enrollment[] {
    query = query.toLowerCase();
    const filteredStudents: Enrollment[] = [];
    enrols.forEach((enrol) => {
        const student = enrol.getUser();
        if (student && foundUser(student, query)) {
            filteredStudents.push(enrol);
        }
    })
    return filteredStudents;
}

export function searchForUsers(users: User[], query: string): User[] {
    query = query.toLowerCase();
    const filteredUsers: User[] = [];
    users.forEach(u => {
        if (foundUser(u, query)) filteredUsers.push(u);
    });
    return filteredUsers;
}

export function searchForGroups(groups: Group[], query: string): Group[] {
    query = query.toLowerCase();
    const filteredGroups: Group[] = [];
    groups.forEach((grp) => {
        if (foundGroup(grp, query)) {
            filteredGroups.push(grp);
        }
    })
    return filteredGroups;
}

export function searchForCourses(courses: Enrollment[] | Course[], query: string): Enrollment[] | Course[] {
    if (courses.length < 1) {
        return courses;
    }
    const enrollmentList: Enrollment[] = [];
    const coursesList: Course[] = [];
    query = query.toLowerCase();
    courses.forEach((e: Enrollment | Course) => {
        const course = e instanceof Enrollment ? e.getCourse() : e;
        if (course && foundCourse(course, query)) {
            e instanceof Enrollment ? enrollmentList.push(e) : coursesList.push(e);
        }
    })
    return enrollmentList.length > 0 ? enrollmentList : coursesList;
}

export function searchForLabs(labs: IAllSubmissionsForEnrollment[], query: string): IAllSubmissionsForEnrollment[] {
    query = query.toLowerCase();
    const filteredLabs: IAllSubmissionsForEnrollment[] = [];
    labs.forEach((e) => {
        const usr = e.enrollment.getUser();
        const grp = e.enrollment.getGroup();
        if (usr && foundUser(usr, query)) {
            filteredLabs.push(e);
        } else if (grp && foundGroup(grp, query)) {
            filteredLabs.push(e);
        }
    });
    return filteredLabs;
}

function foundGroup(group: Group, query: string): boolean {
    return group.getName().toLowerCase().indexOf(query) !== -1
        || group.getTeamid().toString().indexOf(query) !== -1;
}

function foundUser(user: User, query: string): boolean {
    const student = user.toObject();
    return student.name.toLowerCase().indexOf(query) !== -1
        || student.email.toLowerCase().indexOf(query) !== -1
        || student.studentid.toString().indexOf(query) !== -1
        || student.login.toLowerCase().indexOf(query) !== -1;
}

function foundCourse(course: Course, query: string): boolean {
    return course.getName().toLowerCase().indexOf(query) !== -1
        || course.getCode().toLowerCase().indexOf(query) !== -1
        || course.getYear().toString().indexOf(query) !== -1
        || course.getTag().toLowerCase().indexOf(query) !== -1;
}

export function groupRepoLink(groupName: string, courseURL: string): JSX.Element {
    return <a href={courseURL + slugify(groupName)} target="_blank">{groupName}</a>;
}

function gitUserLink(user: string): string {
    return "https://github.com/" + user;
}

function labRepoLink(course: string, login: string): string {
    return course + login + "-labs";
}

// If the courseURL parameter is given, returns a link to the student lab repository,
// otherwise returns link to the user"s GitHub profile.
export function userRepoLink(login: string, name: string, courseURL?: string): JSX.Element {
    return <a href={courseURL ? labRepoLink(courseURL, login) : gitUserLink(login)} target="_blank">{ name }</a>;
}

export function userSubmissionLink(login: string, assignmentName: string, courseURL: string, style?: string): JSX.Element {
    return <a className={style} href={labRepoLink(courseURL, login) + "/tree/master/" + assignmentName} target="_blank">Open repository</a>
}

// Returns a URL-friendly version of the given string.
export function slugify(str: string): string {
    str = str.replace(/^\s+|\s+$/g, "").toLowerCase();

    // Remove accents, swap ñ for n, etc
    const from = "ÁÄÂÀÃÅČÇĆĎÉĚËÈÊẼĔȆÍÌÎÏŇÑÓÖÒÔÕØŘŔŠŤÚŮÜÙÛÝŸŽáäâàãåčçćďéěëèêẽĕȇíìîïňñóöòôõøðřŕšťúůüùûýÿžþÞĐđßÆaæ·/,:;&";
    const to = "AAAAAACCCDEEEEEEEEIIIINNOOOOOORRSTUUUUUYYZaaaaa-cccdeeeeeeeeiiiinnooooo-orrstuuuuuyyzbBDdBAa-------";
    for (let i = 0; i < from.length; i++) {
        str = str.replace(new RegExp(from.charAt(i), "g"), to.charAt(i));
    }

    // Remove invalid chars, replace whitespace by dashes, collapse dashes
    return str.replace(/[^a-z0-9 -_]/g, "").replace(/\s+/g, "-").replace(/-+/g, "-");
}

export function totalScore(reviews: Review[]): number {
    if (reviews.length < 1) return 0;
    let sum = 0;
    reviews.forEach(rv => {
        if (rv.getReady()) {
            sum += rv.getScore();
        }
    });
    return Math.floor(sum / reviews.length);
}

// Some manually graded assignments can have custom max score (not necessary 100%), it will be
// calculated as sum of all scores given for each grading criteria.
export function maxAssignmentScore(benchmarks: GradingBenchmark[]): number {
    let score = 0;
    benchmarks.forEach(bm => {
        bm.getCriteriaList().forEach(c => {
            score += c.getScore();
        });
    });
    return score;
}

export function submissionStatusToString(status?: Submission.Status): string {
    switch (status) {
        case Submission.Status.APPROVED:
            return "Approved";
        case Submission.Status.REJECTED:
            return "Rejected";
        case Submission.Status.REVISION:
            return "Revision";
        default:
            return "None";
    }
}

export function deepCopy(bms: GradingBenchmark[]): GradingBenchmark[] {
    const newList: GradingBenchmark[] = [];
    bms.forEach((bm, i) => {
        const newBm = new GradingBenchmark();
        newBm.setAssignmentid(bm.getAssignmentid());
        newBm.setComment(bm.getComment());
        newBm.setHeading(bm.getHeading());
        newBm.setId(bm.getId());
        const newCriteria: GradingCriterion[] = [];
        bm.getCriteriaList().forEach((c, j) => {
            const newCriterion = new GradingCriterion();
            newCriterion.setId(c.getId());
            newCriterion.setBenchmarkid(c.getBenchmarkid());
            newCriterion.setComment(c.getComment());
            newCriterion.setDescription(c.getDescription());
            newCriterion.setGrade(c.getGrade());
            newCriteria[j] = newCriterion;
        });
        newBm.setCriteriaList(newCriteria);
        newList[i] = newBm;
    });
    return newList;
}

export function setDivider(): JSX.Element {
    return <hr className="list-divider"></hr>;
}

export function hasAllReviews(submission: ISubmission, reviews: number): boolean {
    return submission.reviews.length === reviews;
}

export function submissionStatusSelector(initialStatus: Submission.Status, updateFunc: (status: string) => void, classString?: string): JSX.Element {
    return <div className={"input-group " + classString ?? ""}>
            <span className="input-group-addon">Status: </span>
            <select className="form-control" defaultValue={initialStatus} onChange={(e) => updateFunc(e.target.value)}>
                <option key="st0" value={Submission.Status.NONE} >Set status</option>
                <option key="st1" value={Submission.Status.APPROVED} >Approved</option>
                <option key="st2" value={Submission.Status.REJECTED} >Rejected</option>
                <option key="st3" value={Submission.Status.REVISION} >Revision</option>
            </select>
        </div>;
}

export function mapAllSubmissions(submissions: IAllSubmissionsForEnrollment[], forGroups: boolean, a?: Assignment): Map<(User | Group), ISubmissionLink> {
    const groupMap = new Map<Group, ISubmissionLink>();
    const studentMap = new Map<User, ISubmissionLink>();
    if (!a) {
        return forGroups ? groupMap : studentMap;
    }

    if (forGroups) {
        submissions.forEach(grp => {
            // will return an empty name in case groups stopped preloading on the server side
            // to prevent app crashes
            const group = grp.enrollment.getGroup() ?? new Group();
            let hasSubmission = false;
            grp.labs.forEach(l => {
                if (l.assignment.getId() === a.getId()) {
                    groupMap.set(group, l);
                    hasSubmission = true;
                }
            });
            if (!hasSubmission) {
                groupMap.set(group, {assignment: a, authorName: group.getName()});
            }
        });
        return groupMap;
    }
    submissions.forEach(usr => {
        // will return an empty name in case users stopped preloading on the server side
        // to prevent app crashes
        const user = usr.enrollment.getUser() ?? new User();
        let hasSubmission = false;
        usr.labs.forEach(l => {
            if (l.assignment.getId() === a.getId()) {
                studentMap.set(usr.enrollment.getUser() ?? new User(), l);
                hasSubmission = true;
            }
            if (!hasSubmission) {
                studentMap.set(user, {assignment: a, authorName: user.getName()});
            }
        });
    });
    return studentMap;
}

export function getDaysAfterDeadline(deadline: Date, delivered: Date): number {
    const msInADay = 1000 * 60 * 60 * 24;
    const after =  Math.floor((delivered.valueOf() - deadline.valueOf()) / msInADay);
    return after > 0 ? after : 0;
}

export function isValidUserName(username: string): boolean {
    const onlyCharsAndSpaces = /^[a-zA-Z\s]*$/;
    return onlyCharsAndSpaces.test(username);
}

export function legalIndex(i: number, len: number): boolean {
    return i >= 0 && i <= len - 1;
}
