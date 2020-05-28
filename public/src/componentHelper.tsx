import * as React from "react";
import { Assignment, Course, Enrollment, Group, Review, User, Submission, GradingBenchmark, GradingCriterion } from '../proto/ag_pb';
import { IAllSubmissionsForEnrollment, ISubmissionLink, IStudentLabsForCourse, ISubmission } from './models';

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

export function sortStudentsForRelease<T>(allSubmissions: Map<T, ISubmissionLink>, reviewers: number): T[] {
    const withReviews: T[] = [];
    const withSubmission: T[] = [];
    const noSubmissions: T[] = [];
    allSubmissions.forEach((s, u) => {
        if (s.submission && hasAllReviews(s.submission, reviewers)) {
            console.log(s.authorName + " has all reviews");
            withReviews.push(u);
        } else if (s.submission) {
            console.log(s.authorName + " has submission, no reviews");
            withSubmission.push(u);
        } else {
            console.log(s.authorName + " has no submissions");
            noSubmissions.push(u);
        }
    });
    console.log("with all reviews: " + withReviews.length);
    console.log("with submission: " + withSubmission.length);
    console.log("no submissions: " + noSubmissions.length);
    return withReviews.concat(withSubmission, noSubmissions);
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

export function getSlipDays(allLabs: IStudentLabsForCourse[], selected: ISubmission, forGroups: boolean): number {
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
    return <a className={style} href={labRepoLink(courseURL, login) + "/" + assignmentName} target="_blank">Open repository</a>
}

// Returns a URL-friendly version of the given string.
export function slugify(str: string): string {
    str = str.replace(/^\s+|\s+$/g, "").toLowerCase();

    // Remove accents, swap ñ for n, etc
    const from = "ÁÄÂÀÃÅČÇĆĎÉĚËÈÊẼĔȆÍÌÎÏŇÑÓÖÒÔÕØŘŔŠŤÚŮÜÙÛÝŸŽáäâàãåčçćďéěëèêẽĕȇíìîïňñóöòôõøðřŕšťúůüùûýÿžþÞĐđßÆa·/_,:;";
    const to = "AAAAAACCCDEEEEEEEEIIIINNOOOOOORRSTUUUUUYYZaaaaaacccdeeeeeeeeiiiinnooooooorrstuuuuuyyzbBDdBAa------";
    for (let i = 0; i < from.length; i++) {
        str = str.replace(new RegExp(from.charAt(i), "g"), to.charAt(i));
    }

    // Remove invalid chars, replace whitespace by dashes, collapse dashes
    return str.replace(/[^a-z0-9 -]/g, "").replace(/\s+/g, "-").replace(/-+/g, "-");
}

export function editableListElement(
    text: string,
    defaultText: string,
    toggleFunc: () => void,
    changeFunc: (s: string) => void,
    updateFunc: () => void,
    statebool: boolean
    ): JSX.Element {
    const addDiv = <div className="add-b" onClick={toggleFunc}>{text}</div>;
    const addingDiv = <div className="input-group"><input
        className="form-control m-input"
        type="text"
        defaultValue={defaultText}
        onChange={(e) => changeFunc(e.target.value)}
        onKeyDown={(e) => {
            if (e.key === "Enter") {updateFunc()}
        }}
    />
    <div className="btn-group">
    <button
        className="btn btn-primary btn-xs"
        onClick={updateFunc}>OK</button>
    <button
        className="btn btn-danger btn-xs"
        onClick={toggleFunc}>X</button></div>
    </div>;
    return statebool ? addingDiv : addDiv;
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
