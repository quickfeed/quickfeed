import { isMessage } from "@bufbuild/protobuf"
import React, { memo } from "react"
import { Assignment, Enrollment, EnrollmentSchema, Group, GroupSchema, Submission } from "../../../proto/qf/types_pb"
import { groupRepoLink, isHidden, isTeacher, userRepoLink } from "../../Helpers"
import { useAppState } from "../../overmind"
import SubmissionCell from "./SubmissionCell"

interface MemberRowProps {
    member: Enrollment | Group
    assignments: Assignment[]
    onSubmissionClick: (submission: Submission, owner: Enrollment | Group) => void
    review: boolean
    searchQuery: string
}

/** Represents both possible submissions for a group lab assignment */
interface SubmissionPair {
    individual?: Submission
    group?: Submission
}

const MemberRow = memo(({ member, assignments, onSubmissionClick, review, searchQuery }: MemberRowProps) => {
    const state = useAppState()
    const submissions = state.submissionsForCourse
    const course = state.courses.find(c => c.ID === state.activeCourse)

    const isEnrollment = isMessage(member, EnrollmentSchema)
    const isGroup = isMessage(member, GroupSchema)
    const isMemberTeacher = isEnrollment && isTeacher(member)

    const name = isEnrollment
        ? member.user?.Name ?? "Unknown"
        : member.name

    // Filter by search query
    if (isHidden(name, searchQuery)) {
        return null
    }

    // Generate the repo link for the name cell
    const getRepoLink = (): string | undefined => {
        if (isEnrollment && member.user) {
            return userRepoLink(member.user, course)
        }
        if (isGroup) {
            return groupRepoLink(member, course)
        }
        return undefined
    }

    // Get both individual and group submissions for a group lab assignment
    const getSubmissionPair = (assignment: Assignment): SubmissionPair => {
        if (isGroup) {
            // Groups only have group submissions
            const group = submissions.ForGroup(member)?.find(s => s.AssignmentID === assignment.ID)
            return group ? { group } : {}
        }
        if (isEnrollment) {
            // Find submission for this assignment
            const userSubmission = submissions.ForUser(member)?.find(s => s.AssignmentID === assignment.ID)

            // An individual submission has groupID === 0
            // A group submission has groupID > 0
            const isIndividualSubmission = userSubmission && userSubmission.groupID === 0n
            const individual = isIndividualSubmission ? userSubmission : undefined

            // For group labs, also check for group submission using ForGroup with enrollment
            if (assignment.isGroupLab && member.groupID > 0n) {
                const group = submissions.ForGroup(member)?.find(s => s.AssignmentID === assignment.ID)
                // Only include properties that actually have submissions
                const result: SubmissionPair = {}
                if (individual) result.individual = individual
                if (group) result.group = group
                return result
            }

            return individual ? { individual } : {}
        }
        return {}
    }

    const repoLink = getRepoLink()

    return (
        <tr>
            <th className="font-medium">
                <div className="flex items-center gap-2">
                    <MemberName name={name} repoLink={repoLink} />
                    {isMemberTeacher && (
                        <span className="badge badge-primary badge-sm">Teacher</span>
                    )}
                </div>
            </th>
            {state.isCourseManuallyGraded && (
                <td className="text-base-content/70">
                    {isEnrollment ? member.userID.toString() : member.ID.toString()}
                </td>
            )}
            {assignments.map(assignment => {
                const pair = getSubmissionPair(assignment)
                const hasNoSubmission = !pair.individual && !pair.group

                if (hasNoSubmission) {
                    return (
                        <td key={assignment.ID.toString()} className="text-base-content/50 text-center">
                            —
                        </td>
                    )
                }

                return (
                    <SubmissionCell
                        key={assignment.ID.toString()}
                        submissionPair={pair}
                        owner={member}
                        onSubmissionClick={(submission) => onSubmissionClick(submission, member)}
                        review={review}
                    />
                )
            })}
        </tr>
    )
})

MemberRow.displayName = "MemberRow"

/** Renders the member name, optionally as a link */
const MemberName = ({ name, repoLink }: { name: string; repoLink?: string }) => {
    if (repoLink) {
        return (
            <a
                href={repoLink}
                target="_blank"
                rel="noopener noreferrer"
                className="link link-hover"
            >
                {name}
            </a>
        )
    }
    return <span>{name}</span>
}

export default MemberRow
