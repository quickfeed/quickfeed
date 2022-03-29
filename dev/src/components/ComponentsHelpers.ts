import { Assignment, Enrollment, SubmissionLink } from "../../proto/ag/ag_pb"
import { groupRepoLink, userLink, userRepoLink } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { UserCourseSubmissions } from "../overmind/state"
import { RowElement, Row } from "./DynamicTable"


export const generateSubmissionRows = (links: UserCourseSubmissions[], cellGenerator: (s: SubmissionLink, e?: Enrollment) => RowElement, groupName?: boolean, assignmentID?: number): Row[] => {
    const state = useAppState()
    const course = state.courses.find(c => c.getId() === state.activeCourse)
    return links?.map((link) => {
        const row: Row = []
        if (link.enrollment && link.user) {
            const url = course ? userRepoLink(course, link.user) : userLink(link.user)
            row.push({ value: link.user.getName(), link: url })
            groupName && row.push(link.enrollment.getGroup()?.getName() ?? "")
        } else if (link.group) {
            const data: RowElement = course ? { value: link.group.getName(), link: groupRepoLink(course, link.group) } : link.group.getName()
            row.push(data)
        }
        if (link.submissions) {
            for (const submissionLink of link.submissions) {
                if (state.review.assignmentID > 0 && submissionLink.getAssignment()?.getId() != state.review.assignmentID) {
                    continue
                }
                row.push(cellGenerator(submissionLink, link.enrollment))
            }
        }
        return row
    })
}

export const generateAssignmentsHeader = (base: RowElement[], assignments: Assignment[], group: boolean, assignmentID?: number): Row => {
    const actions = useActions()
    for (const assignment of assignments) {
        if (assignmentID && assignment.getId() !== assignmentID) {
            continue
        }
        if (group && assignment.getIsgrouplab()) {
            base.push({ value: `${assignment.getName()} (g)`, onClick: () => actions.review.setAssignmentID(assignment.getId()) })
        }
        if (!group) {
            base.push({ value: assignment.getIsgrouplab() ? `${assignment.getName()} (g)` : assignment.getName(), onClick: () => actions.review.setAssignmentID(assignment.getId()) })
        }
    }
    return base
}
