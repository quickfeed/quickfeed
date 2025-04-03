import { isMessage } from "@bufbuild/protobuf"
import { Assignment, Course, Enrollment, EnrollmentSchema, Group, GroupSchema, Submission } from "../../proto/qf/types_pb"
import { groupRepoLink, Icon, SubmissionsForCourse, SubmissionSort, userRepoLink } from "../Helpers"
import { useActions } from "../overmind"
import { AssignmentsMap, State } from "../overmind/state"
import { CellElement, Row, RowElement } from "./DynamicTable"
import { Icons } from "./Icons"

export const generateSubmissionRows = (elements: Enrollment[] | Group[], generator: (s: Submission, e?: Enrollment | Group) => RowElement, state: State): Row[] => {
    const course = state.courses.find(c => c.ID === state.activeCourse)
    const assignments = state.getAssignmentsMap(state.activeCourse)
    return elements.map(element => {
        return generateRow(element, assignments, state.submissionsForCourse, generator, state.individualSubmissionView, course, state.isCourseManuallyGraded)
    })
}

export const generateRow = (
    enrollment: Enrollment | Group,
    assignments: AssignmentsMap,
    submissions: SubmissionsForCourse,
    generator: (s: Submission, e?: Enrollment | Group) => RowElement,
    individual: boolean,
    course?: Course,
    withID?: boolean
): Row => {
    const row: Row = []
    const isEnrollment = isMessage(enrollment, EnrollmentSchema)
    const isGroup = isMessage(enrollment, GroupSchema)

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

        if (isGroup && isGroupLab) {
            // If we're dealing with a group assignment and a group, we should try to find a group submission
            submission = submissions.ForGroup(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
        }

        if (isEnrollment) {
            if (isGroupLab && enrollment.groupID === 0n) {
                // If we're dealing with a group assignment, and the enrollment is not part of a group
                // we should try to find an individual submission instead
                submission = submissions.ForUser(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
            } else if (isGroupLab && !individual) {
                // If we're dealing with a group assignment, and the user is not viewing individual submissions
                submission = submissions.ForGroup(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
            } else {
                submission = submissions.ForUser(enrollment)?.find(s => s.AssignmentID.toString() === assignmentID)
            }
        }

        if (submission) {
            row.push(generator(submission, enrollment))
            return
        }
        row.push(Icons.GreyDash)
    })
    return row
}

export const generateAssignmentsHeader = (assignments: Assignment[], viewByGroup: boolean, actions: ReturnType<typeof useActions>, isCourseManuallyGraded: boolean): Row => {
    const base: Row = [
        { value: "Name", onClick: () => actions.setSubmissionSort(SubmissionSort.Name) }
    ]
    if (isCourseManuallyGraded) {
        base.unshift({ value: "ID", onClick: () => actions.setSubmissionSort(SubmissionSort.ID) })
    }
    for (const assignment of assignments) {
        const cell: CellElement = { value: assignment.name, onClick: () => actions.review.setAssignmentID(assignment.ID) }
        // If we are viewing by group, ignore all non-group assignments
        if (viewByGroup && !assignment.isGroupLab) {
            continue
        }
        // If group assignment, add group icon
        if (assignment.isGroupLab) {
            cell.iconTitle = "Group"; cell.iconClassName = Icon.GROUP
        } else {
            cell.iconTitle = "Individual"; cell.iconClassName = Icon.USER
        }
        base.push(cell)
    }
    return base
}
