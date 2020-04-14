import * as React from "react";
import { Enrollment } from "../proto/ag_pb";


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
        }
    })
    sorted = sorted.concat(active, archived);
    return sorted;
}