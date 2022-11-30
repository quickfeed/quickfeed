import { Assignment, Enrollment, Group, Submission } from "../../proto/qf/types_pb"
import { groupRepoLink, userLink, userRepoLink } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { RowElement, Row } from "./DynamicTable"

export const generateSubmissionRows = (elements: Enrollment[] | Group[], generator: (s: Submission, e?: Enrollment | Group) => RowElement, assignmentIDs: bigint[], groupName?: boolean): Row[] => {
    const state = useAppState()
    const submissions = state.groupView ? state.submissionsByGroup : state.submissionsByEnrollment
    const course = state.courses.find(c => c.ID === state.activeCourse)
    return elements.map(element => {
        const row: Row = []
        if (element instanceof Enrollment && element.user) {
            const url = course ? userRepoLink(course, element.user) : userLink(element.user)
            row.push({ value: element.user.Name, link: url })
            groupName && row.push(element.group?.name ?? "")
        } else if (element instanceof Group) {
            const data: RowElement = course ? { value: element.name, link: groupRepoLink(course, element) } : element.name
            row.push(data)
        }
        const submissionsList = submissions.get(element.ID)
        assignmentIDs?.forEach(assignment => {
            const submission = submissionsList?.find(s => s.AssignmentID === assignment)
            if (submission) {
                row.push(generator(submission, element))
            } else {
                row.push("N/A")
            }
        })

        return row
    })
}

export const generateAssignmentsHeader = (base: RowElement[], assignments: Assignment[], group: boolean, assignmentID?: bigint): Row => {
    const actions = useActions()
    for (const assignment of assignments) {
        if (assignmentID && assignment.ID !== assignmentID) {
            continue
        }
        if (group && assignment.isGroupLab) {
            base.push({ value: `${assignment.name} (g)`, onClick: () => actions.review.setAssignmentID(assignment.ID) })
        }
        if (!group) {
            base.push({ value: assignment.isGroupLab ? `${assignment.name} (g)` : assignment.name, onClick: () => actions.review.setAssignmentID(assignment.ID) })
        }
    }
    return base
}
