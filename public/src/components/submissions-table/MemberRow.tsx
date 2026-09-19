import { isMessage } from "@bufbuild/protobuf"
import { memo } from "react"
import type { Assignment, Enrollment, Group, Submission } from "../../../proto/qf/types_pb"
import { EnrollmentSchema, GroupSchema } from "../../../proto/qf/types_pb"
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
                if (individual) { result.individual = individual }
                if (group) { result.group = group }
                return result
            }

            return individual ? { individual } : {}
        }
        return {}
    }

    const repoLink = getRepoLink()

    // The GitHub login is the name of the student's repository, and the group name is
    // the name of the group repository; showing both saves the teacher from hovering
    // the name link to find out which repository belongs to which student.
    // Group rows already show the group (repository) name as their name.
    const login = isEnrollment ? member.user?.Login : undefined
    const groupName = isEnrollment ? enrollmentGroupName(member, state.groups[state.activeCourse.toString()]) : undefined

    return (
        <tr>
            <th className="font-medium">
                <div className="flex items-center gap-2">
                    <MemberName name={name} repoLink={repoLink} />
                    {isMemberTeacher && (
                        <span className="badge badge-primary badge-sm">Teacher</span>
                    )}
                </div>
                <MemberDetails login={login} groupName={groupName} />
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

/** Returns the name of the group the enrollment belongs to, if any.
 *  The group is normally preloaded on the enrollment for teachers; the course
 *  groups are used as a fallback in case it is not. */
const enrollmentGroupName = (enrollment: Enrollment, groups?: Group[]): string | undefined => {
    if (enrollment.groupID === 0n) {
        return undefined
    }
    return enrollment.group?.name ?? groups?.find(g => g.ID === enrollment.groupID)?.name
}

/** Renders the member's GitHub login and group name below the member name */
const MemberDetails = ({ login, groupName }: { login?: string; groupName?: string }) => {
    if (!login && !groupName) {
        return null
    }
    return (
        <div className="text-base-content/60 text-xs font-normal">
            {login ? <span>@{login}</span> : null}
            {login && groupName ? <span className="mx-1">·</span> : null}
            {groupName ? <span>{groupName}</span> : null}
        </div>
    )
}

/** Renders the member name, optionally as a link */
const MemberName = ({ name, repoLink }: { name: string; repoLink?: string }) => {
    if (repoLink) {
        return (
            <a href={repoLink} target="_blank" rel="noopener noreferrer" className="link link-hover">
                {name}
            </a>
        )
    }
    return <span>{name}</span>
}

export default MemberRow
