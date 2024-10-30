import { Assignment, Course, Enrollment, Group, Submission } from "../../proto/qf/types_pb"
import { groupRepoLink, SubmissionsForCourse, SubmissionSort, userRepoLink } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { AssignmentsMap } from "../overmind/state"
import { Row, RowElement } from "./DynamicTable"


export const generateSubmissionRows = (elements: Enrollment[] | Group[], generator: (s: Submission, e?: Enrollment | Group) => RowElement): Row[] => {
    const state = useAppState()
    const course = state.courses.find(c => c.ID === state.activeCourse)
    const assignments = state.getAssignmentsMap(state.activeCourse)
    return elements.map(element => {
        return generateRow(element, assignments, state.submissionsForCourse, generator, course, state.isCourseManuallyGraded)
    })
}

export const generateRow = (enrollment: Enrollment | Group, assignments: AssignmentsMap, submissions: SubmissionsForCourse, generator: (s: Submission, e?: Enrollment | Group) => RowElement, course?: Course, withID?: boolean): Row => {
    const row: Row = []
    const isEnrollment = enrollment.$typeName === "qf.Enrollment"
    const isGroup = enrollment.$typeName === "qf.Group"

    if (withID) {
        isEnrollment
            ? row.push({ value: enrollment.userID.toString() })
            : row.push({ value: enrollment.ID.toString() })
    }

    if (isEnrollment && enrollment.user) {
        row.push({ value: enrollment.user.Name, link: userRepoLink(enrollment.user, course) })
    } else if (isGroup) {
        row.push({ value: enrollment.name, link: groupRepoLink(enrollment, course) })
    }

    Object.entries(assignments)?.forEach(([assignmentID, isGroupLab]) => {
        let submission: Submission | undefined
        if (isGroup && !isGroupLab) {
            // If the assignment is not a group assignment, and we're dealing with a group
            // we should exit early without adding to the row.
            return
        }

        if (isGroupLab && isEnrollment && enrollment.groupID === 0n) {
            // If we're dealing with a group assignment, and the enrollment is not part of a group
            // we should try to find an individual submission instead
            submission = submissions.ForUser(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
        } else if (isGroupLab) {
            // If the previous conditions are not met, we have this situation:
            // - The assignment is a group assignment
            // - We're either dealing with an enrollment that is part of a group
            // - or we're dealing with a group
            submission = submissions.ForGroup(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
        } else if (isEnrollment) {
            submission = submissions.ForUser(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
        }
        if (submission) {
            row.push(generator(submission, enrollment))
            return
        }
        row.push("N/A")
    })
    return row
}

export const generateAssignmentsHeader = (assignments: Assignment[], group: boolean): Row => {
    const isCourseManuallyGraded = useAppState((state) => state.isCourseManuallyGraded)
    const actions = useActions()
    const base: Row = [
        { value: "Name", onClick: () => actions.setSubmissionSort(SubmissionSort.Name) }
    ]
    if (isCourseManuallyGraded) {
        base.unshift({ value: "ID", onClick: () => actions.setSubmissionSort(SubmissionSort.ID) })
    }
    for (const assignment of assignments) {
        if (group && assignment.isGroupLab) {
            base.push({ value: `${assignment.name} (g)`, onClick: () => actions.review.setAssignmentID(assignment.ID) })
        }
        if (!group) {
            base.push({ value: assignment.isGroupLab ? `${assignment.name} (g)` : assignment.name, onClick: () => actions.review.setAssignmentID(assignment.ID) })
        }
    }
    return base
}
