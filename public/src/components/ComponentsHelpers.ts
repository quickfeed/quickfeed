import { Assignment, Enrollment } from "../../proto/qf/types/types_pb"
import { SubmissionLink } from "../../proto/qf/types/requests_pb"
import { groupRepoLink, userLink, userRepoLink } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { UserCourseSubmissions } from "../overmind/state"
import { RowElement, Row } from "./DynamicTable"

export const generateSubmissionRows = (links: UserCourseSubmissions[], cellGenerator: (s: SubmissionLink.AsObject, e?: Enrollment.AsObject) => RowElement, groupName?: boolean): Row[] => {
    const state = useAppState()
    const course = state.courses.find(c => c.id === state.activeCourse)
    return links?.map((link) => {
        const row: Row = []
        if (link.enrollment && link.user) {
            const url = course ? userRepoLink(course, link.user) : userLink(link.user)
            row.push({ value: link.user.name, link: url })
            groupName && row.push(link.enrollment.group?.name ?? "")
        } else if (link.group) {
            const data: RowElement = course ? { value: link.group.name, link: groupRepoLink(course, link.group) } : link.group.name
            row.push(data)
        }
        if (link.submissions) {
            for (const submissionLink of link.submissions) {
                if (state.review.assignmentID > 0 && submissionLink.assignment?.id != state.review.assignmentID) {
                    continue
                }
                row.push(cellGenerator(submissionLink, link.enrollment))
            }
        }
        return row
    })
}

export const generateAssignmentsHeader = (base: RowElement[], assignments: Assignment.AsObject[], group: boolean, assignmentID?: number): Row => {
    const actions = useActions()
    for (const assignment of assignments) {
        if (assignmentID && assignment.id !== assignmentID) {
            continue
        }
        if (group && assignment.isgrouplab) {
            base.push({ value: `${assignment.name} (g)`, onClick: () => actions.review.setAssignmentID(assignment.id) })
        }
        if (!group) {
            base.push({ value: assignment.isgrouplab ? `${assignment.name} (g)` : assignment.name, onClick: () => actions.review.setAssignmentID(assignment.id) })
        }
    }
    return base
}
