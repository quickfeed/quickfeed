import { Assignment, Submission } from "../../../proto/ag_pb";
import { IAllSubmissionsForEnrollment, ISubmissionLink } from "../../models";

export function sortByScore(students: IAllSubmissionsForEnrollment[], labs: Assignment[], isGroupLab: boolean): IAllSubmissionsForEnrollment[] {
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
    const withSubmission: IAllSubmissionsForEnrollment[] = [];
    const withoutSubmission: IAllSubmissionsForEnrollment[] = [];
    // split all students into two arrays: with and without submission to the last lab
    students.forEach((ele) => {
        let hasSubmission = false;
        ele.labs.forEach((a) => {
            // check if there is a submission for the latest course assignment
            if (a.assignment.getId() === assignmentID && a.submission) {
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
        const leftLab = left.labs[assignmentIndex].submission;
        const rightLab = right.labs[assignmentIndex].submission;
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

export function generateCellClass(lab: ISubmissionLink): string {
    if (lab.submission && lab.submission.status === Submission.Status.APPROVED) {
        return "approved-cell";
    }
    const passing = ((lab.assignment.getScorelimit() > 0)
     && lab.submission && (lab.submission.score >= lab.assignment.getScorelimit()));
    return passing ? "passing" : "";
}