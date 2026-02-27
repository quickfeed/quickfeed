import React, { memo, useCallback } from "react"
import { Assignment, Enrollment, Group, Submission } from "../../../proto/qf/types_pb"
import { Icon, SubmissionSort } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import MemberRow from "./MemberRow"

interface SubmissionsTableProps {
    onSubmissionClick: (submission: Submission, owner: Enrollment | Group) => void
    review?: boolean
}

const SubmissionsTable = memo(({ onSubmissionClick, review = false }: SubmissionsTableProps) => {
    const state = useAppState()
    const actions = useActions()

    const { courseMembers, activeCourse, groupView, query, review: reviewState, isCourseManuallyGraded } = state

    const assignments = state.assignments[activeCourse.toString()]?.filter(
        a => reviewState.assignmentID <= 0 || a.ID === reviewState.assignmentID
    ) ?? []

    // Filter assignments by group/individual view
    const filteredAssignments = groupView
        ? assignments.filter(a => a.isGroupLab)
        : assignments

    const handleSort = useCallback((sortBy: SubmissionSort) => () => {
        actions.global.setSubmissionSort(sortBy)
    }, [actions])

    const handleAssignmentClick = useCallback((assignment: Assignment) => () => {
        actions.review.setAssignmentID(assignment.ID)
    }, [actions])

    if (courseMembers.length === 0) {
        return (
            <div className="text-center py-8 text-base-content/70">
                No submissions to display
            </div>
        )
    }

    return (
        <table className="table table-zebra table-pin-rows">
            <thead className="bg-base-300">
                <tr>
                    <th
                        className="cursor-pointer hover:bg-base-200"
                        onClick={handleSort(SubmissionSort.Name)}
                    >
                        Name
                    </th>
                    {isCourseManuallyGraded && (
                        <th
                            className="cursor-pointer hover:bg-base-200"
                            onClick={handleSort(SubmissionSort.ID)}
                        >
                            ID
                        </th>
                    )}
                    {filteredAssignments.map(a => (
                        <th
                            key={a.ID.toString()}
                            className="cursor-pointer hover:bg-base-200"
                            onClick={handleAssignmentClick(a)}
                        >
                            <div className="flex items-center gap-1">
                                <span>{a.name}</span>
                                <i
                                    className={a.isGroupLab ? Icon.GROUP : Icon.USER}
                                    title={a.isGroupLab ? "Group" : "Individual"}
                                />
                            </div>
                        </th>
                    ))}
                </tr>
            </thead>
            <tbody>
                {courseMembers.map(member => (
                    <MemberRow
                        key={getMemberKey(member)}
                        member={member}
                        assignments={filteredAssignments}
                        onSubmissionClick={onSubmissionClick}
                        review={review}
                        searchQuery={query}
                    />
                ))}
            </tbody>
        </table>
    )
})

// Helper to get a unique key for each member
const getMemberKey = (member: Enrollment | Group): string => {
    if ('userID' in member) {
        return `enrollment-${member.ID.toString()}`
    }
    return `group-${member.ID.toString()}`
}

SubmissionsTable.displayName = "SubmissionsTable"

export default SubmissionsTable
