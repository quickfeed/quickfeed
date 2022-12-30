import { Assignment, Course, Enrollment, Group, Submission } from "../../proto/qf/types_pb"
import { groupRepoLink, SubmissionsForCourse, userRepoLink } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { AssignmentsMap } from "../overmind/state"
import { RowElement, Row } from "./DynamicTable"


export const generateSubmissionRows = (elements: Enrollment[] | Group[], generator: (s: Submission, e?: Enrollment | Group) => RowElement): Row[] => {
    const state = useAppState()
    const course = state.courses.find(c => c.ID === state.activeCourse)
    const assignments = state.getAssignmentsMap(state.activeCourse)
    return elements.map(element => {
        return generateRow(element, assignments, state.submissionsForCourse, generator, course)
    })
}

const generateRow = (enrollment: Enrollment | Group, assignments: AssignmentsMap, submissions: SubmissionsForCourse, generator: (s: Submission, e?: Enrollment | Group) => RowElement, course?: Course): Row => {
    const row: Row = []
    const isEnrollment = enrollment instanceof Enrollment
    const isGroup = enrollment instanceof Group

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
        if (isGroupLab) {
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

export const generateAssignmentsHeader = (base: RowElement[], assignments: Assignment[], group: boolean): Row => {
    const actions = useActions()
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
