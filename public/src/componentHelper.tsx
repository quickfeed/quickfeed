import * as React from "react";
import { Course, Enrollment } from "../proto/ag_pb";


export function sortCoursesByVisibility(enrols: Enrollment[]): Enrollment[] {
    let sorted: Enrollment[] = [];
    const active: Enrollment[] = [];
    const archived: Enrollment[] = [];
    // TODO: if we want to display active and hidden courses in separate tables,
    // they can be easily separated and set as a new state here
    enrols.forEach((enrol) => {
        switch (enrol.getState()) {
            case Enrollment.DisplayState.FAVORITE:
                sorted.push(enrol);
                break;
            case Enrollment.DisplayState.ACTIVE:
                active.push(enrol);
                break;
            case Enrollment.DisplayState.ARCHIVED:
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

export function searchForCourses(enrols: Enrollment[], query: string): Enrollment[] {
    const filteredCourses: Enrollment[] = [];
    enrols.forEach((enrol) => {
        const course = enrol.getCourse();
        if (course) {
            if (course.getName().toLowerCase().indexOf(query) !== -1 ||
                course.getCode().toLowerCase().indexOf(query) !== -1 ||
                course.getYear().toString().indexOf(query) !== -1 ||
                course.getTag().toLowerCase().indexOf(query) !== -1) {
                    filteredCourses.push(enrol);
                }
        }
    });
    return filteredCourses;
}

export function getActiveCourses(courses: Course[], enrols: Enrollment[], userID: number): Course[] {
    const activeCourses: Course[] = [];
    enrols.forEach((enrol) => {
        const crs = enrol.getCourse();
        if (enrol.getState() !== Enrollment.DisplayState.ARCHIVED &&
            crs && courses.find(e => e.getId() === crs.getId()
            )) {
            activeCourses.push(crs);
        }
    });
    return activeCourses;
}

export function groupRepoLink(groupName: string, courseURL: string): JSX.Element {
    return <a href={courseURL + slugify(groupName)} target="_blank">{ groupName }</a>;
}

function gitUserLink(user: string): string {
    return "https://github.com/" + user;
}

function labRepoLink(course: string, login: string): string {
    return course + login + "-labs";
}

// If the courseURL parameter is given, returns a link to the student lab repository,
// otherwise returns link to the user's GitHub profile.
export function generateGitLink(user: string, courseURL?: string): string {
    return courseURL ? labRepoLink(courseURL, user) : gitUserLink(user);
}

// Returns a URL-friendly version of the given string.
export function slugify(str: string): string {

    str = str.replace(/^\s+|\s+$/g, "").toLowerCase();

    // Remove accents, swap ñ for n, etc
    const from = "ÁÄÂÀÃÅČÇĆĎÉĚËÈÊẼĔȆÍÌÎÏŇÑÓÖÒÔÕØŘŔŠŤÚŮÜÙÛÝŸŽáäâàãåčçćďéěëèêẽĕȇíìîïňñóöòôõøðřŕšťúůüùûýÿžþÞĐđßÆa·/_,:;";
    const to   = "AAAAAACCCDEEEEEEEEIIIINNOOOOOORRSTUUUUUYYZaaaaaacccdeeeeeeeeiiiinnooooooorrstuuuuuyyzbBDdBAa------";
    for (let i = 0 ; i < from.length ; i++) {
        str = str.replace(new RegExp(from.charAt(i), "g"), to.charAt(i));
    }

    // Remove invalid chars, replace whitespace by dashes, collapse dashes
    str = str.replace(/[^a-z0-9 -]/g, "").replace(/\s+/g, "-").replace(/-+/g, "-");

    return str;
}
