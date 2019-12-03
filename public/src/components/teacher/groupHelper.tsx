import * as React from "react";
import { Assignment } from "../../../proto/ag_pb";
import { IAssignmentLink } from "../../models";

export function sortByScore(students: IAssignmentLink[], labs: Assignment[], isGroupLab: boolean): IAssignmentLink[] {
    // if no assignments yet, disregard
    if (labs.length < 1) {
        return students;
    }
    const allLabs = labs.slice().reverse();
    // find the latest individual assignment and its index
    let assignmentID = 0;
    let assignmentIndex = 0;
    let latestLab = null;
    if (isGroupLab) {
        latestLab = allLabs.find((lab) => {
            return lab.getIsgrouplab();
        });
    } else {
        latestLab = allLabs.find((lab) => {
            return !lab.getIsgrouplab();
        });
    }

    if (latestLab) {
        assignmentID = latestLab.getId();
        assignmentIndex = labs.indexOf(latestLab);
    }
    const withSubmission: IAssignmentLink[] = [];
    const withoutSubmission: IAssignmentLink[] = [];
    // split all students into two arrays: with and without submission to the last lab
    students.forEach((ele) => {
        let hasSubmission = false;
        ele.assignments.forEach((a) => {
            // check if there is a submission for the latest course assignment
            if (a.assignment.getId() === assignmentID && a.latest) {
                hasSubmission = true;
            }
        });
        if (hasSubmission) {
            withSubmission.push(ele);
        } else {
            withoutSubmission.push(ele);
        }
    });
    // sort students with submissions
    const sorted = withSubmission.sort((left, right) => {
        const leftLab = left.assignments[assignmentIndex].latest;
        const rightLab = right.assignments[assignmentIndex].latest;
        if (leftLab && rightLab) {
            if (leftLab.score > rightLab.score) {
                return -1;
            } else if (leftLab.score < rightLab.score) {
                return 1;
            } else {
                return 0;
            }
        }
        return 0;
    });
    // then add students without submission at the end of list
    const fullList = sorted.concat(withoutSubmission);
    return fullList;
}

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

export function generateGroupRepoLink(groupName: string, courseURL: string): JSX.Element {
    return <a href={courseURL + slugify(groupName)} target="_blank">{ groupName }</a>;
}
