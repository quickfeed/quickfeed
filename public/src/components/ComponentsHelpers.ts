import { Assignment, Enrollment, Group, Submission } from "../../proto/qf/types_pb"
import { groupRepoLink, userRepoLink } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { RowElement, Row } from "./DynamicTable"

export const generateSubmissionRows = (elements: Enrollment[] | Group[], generator: (s: Submission, e?: Enrollment | Group) => RowElement, assignmentIDs: bigint[]): Row[] => {
    const state = useAppState()
    const submissions = state.groupView ? state.submissionsByGroup : state.submissionsByEnrollment
    const course = state.courses.find(c => c.ID === state.activeCourse)
    return elements.map(element => {
        const row: Row = []
        if (element instanceof Enrollment && element.user) {
            row.push({ value: element.user.Name, link: userRepoLink(element.user, course) })
        } else if (element instanceof Group) {
            row.push({ value: element.name, link: groupRepoLink(element, course) })
        }
        const submissionsList = submissions[element.ID.toString()].submissions
        assignmentIDs?.forEach(assignment => {
            const submission = submissionsList?.find(s => s.AssignmentID === assignment)
            if (submission) {
                row.push(generator(submission, element))
                return
            }
            row.push("N/A")
        })
        return row
    })
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
