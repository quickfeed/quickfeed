import * as React from "react";
import { Course, Enrollment, Group, User } from "../proto/ag_pb";
import { IStudentLabsForCourse } from "./models";

export function sortEnrollmentsByVisibility(enrols: Enrollment[]): Enrollment[] {
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
                archived.push(enrol);
                break;
            case Enrollment.DisplayState.UNSET:
                active.push(enrol);
                break;
        }
    })
    sorted = sorted.concat(active, archived);
    return sorted;
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
            default:
                break;
        }
    });
    favorite = favorite.concat(active);
    return favorite;
}

export function sortUsersByAdminStatus(users: Enrollment[]): Enrollment[] {
    return users.sort((x, y) => ((x.getUser()?.getIsadmin() ?? false) < (y.getUser()?.getIsadmin() ?? false) ? 1 : -1));
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

export function searchForLabs(labs: IStudentLabsForCourse[], query: string): IStudentLabsForCourse[] {
    query = query.toLowerCase();
    const filteredLabs: IStudentLabsForCourse[] = [];
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
// otherwise returns link to the user's GitHub profile.
export function userRepoLink(login: string, name: string, courseURL?: string): JSX.Element {
    return <a href={courseURL ? labRepoLink(courseURL, login) : gitUserLink(login)} target="_blank">{ name }</a>;
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
