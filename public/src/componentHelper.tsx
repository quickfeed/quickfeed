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