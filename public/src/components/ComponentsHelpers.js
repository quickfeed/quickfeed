import { isMessage } from "@bufbuild/protobuf";
import { EnrollmentSchema, GroupSchema } from "../../proto/qf/types_pb";
import { groupRepoLink, Icon, SubmissionSort, userRepoLink } from "../Helpers";
import { Icons } from "./Icons";
export const generateSubmissionRows = (elements, generator, state) => {
    const course = state.courses.find(c => c.ID === state.activeCourse);
    const assignments = state.getAssignmentsMap(state.activeCourse);
    return elements.map(element => {
        return generateRow(element, assignments, state.submissionsForCourse, generator, state.individualSubmissionView, course, state.isCourseManuallyGraded);
    });
};
export const generateRow = (enrollment, assignments, submissions, generator, individual, course, withID) => {
    const row = [];
    const isEnrollment = isMessage(enrollment, EnrollmentSchema);
    const isGroup = isMessage(enrollment, GroupSchema);
    if (withID) {
        const ID = isEnrollment ? enrollment.userID : enrollment.ID;
        row.push({ value: ID.toString() });
    }
    if (isEnrollment && enrollment.user) {
        row.push({ value: enrollment.user.Name, link: userRepoLink(enrollment.user, course) });
    }
    else if (isGroup) {
        row.push({ value: enrollment.name, link: groupRepoLink(enrollment, course) });
    }
    Object.entries(assignments)?.forEach(([assignmentID, isGroupLab]) => {
        let submission;
        if (isGroup && !isGroupLab) {
            return;
        }
        if (isGroup && isGroupLab) {
            submission = submissions.ForGroup(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID);
        }
        if (isEnrollment) {
            if (isGroupLab && enrollment.groupID === 0n) {
                submission = submissions.ForUser(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID);
            }
            else if (isGroupLab && !individual) {
                submission = submissions.ForGroup(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID);
            }
            else {
                submission = submissions.ForUser(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID);
            }
        }
        if (submission) {
            row.push(generator(submission, enrollment));
            return;
        }
        row.push(Icons.NotAvailable);
    });
    return row;
};
export const generateAssignmentsHeader = (assignments, viewByGroup, actions, isCourseManuallyGraded) => {
    const handleSort = (sortBy) => () => actions.global.setSubmissionSort(sortBy);
    const base = [
        { value: "Name", onClick: handleSort(SubmissionSort.Name) }
    ];
    if (isCourseManuallyGraded) {
        base.unshift({ value: "ID", onClick: handleSort(SubmissionSort.ID) });
    }
    for (const assignment of assignments) {
        const cell = { value: assignment.name, onClick: () => actions.review.setAssignmentID(assignment.ID) };
        if (viewByGroup && !assignment.isGroupLab) {
            continue;
        }
        if (assignment.isGroupLab) {
            cell.iconTitle = "Group";
            cell.iconClassName = Icon.GROUP;
        }
        else {
            cell.iconTitle = "Individual";
            cell.iconClassName = Icon.USER;
        }
        base.push(cell);
    }
    return base;
};
