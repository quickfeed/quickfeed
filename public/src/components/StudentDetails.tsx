import React, { useEffect } from "react"
import { Link } from "react-router-dom"
import type { Assignment, Note, Submission } from "../../proto/qf/types_pb"
import { Submission_Status } from "../../proto/qf/types_pb"
import { EnrollmentStatus, getFormattedTime, getStatusByUser, SubmissionStatus, submissionStatusConfig, userRepoLink } from "../Helpers"
import { useCourseID } from "../hooks/useCourseID"
import { useEnrollmentID } from "../hooks/useEnrollmentID"
import { useActions, useAppState } from "../overmind"
import Avatar from "./Avatar"
import Badge from "./Badge"
import type { LabelledTarget, TargetInfo } from "./Notes"
import { NotePanelBody } from "./Notes"

/**
 * StudentDetails is a teacher-only overview of a single student in a course:
 * their enrollment info, slip-days, submissions, and internal staff notes.
 * It is reached from the members page at /course/:id/members/:enrollmentID.
 */
const StudentDetails = () => {
    const state = useAppState()
    const actions = useActions()
    const courseID = useCourseID()
    const id = useEnrollmentID()
    const root = `/course/${courseID}`

    useEffect(() => {
        // Ensure submissions and notes are loaded even on direct navigation.
        if (!state.loadedCourse[courseID.toString()]) {
            actions.global.loadCourseSubmissions(courseID)
        }
        actions.notes.getCourseNotes()
    }, [actions, courseID, state.loadedCourse])

    const enrollment = (state.courseEnrollments[courseID.toString()] ?? []).find(e => e.ID === id)
    if (!enrollment?.user) {
        return (
            <div className="p-6">
                <Link to={`${root}/members`} className="link link-hover text-primary">← Back to members</Link>
                <p className="mt-4 text-base-content/60">Student not found.</p>
            </div>
        )
    }

    const user = enrollment.user
    const course = state.courses.find(c => c.ID === courseID)
    const assignments = state.assignments[courseID.toString()] ?? []
    const submissions = state.submissionsForCourse.ForUser(enrollment)
    const groups = state.groups[courseID.toString()] ?? []

    // Notes shown for the student: those attached directly to the enrollment plus the student's group notes.
    const notes = state.notes.courseNotes.filter(n =>
        n.EnrollmentID === enrollment.ID || (enrollment.groupID > 0n && n.GroupID === enrollment.groupID)
    )
    const targets: LabelledTarget[] = [{ label: "Student", value: { EnrollmentID: enrollment.ID } }]
    if (enrollment.groupID > 0n) {
        targets.push({ label: enrollment.group?.name ? `Group: ${enrollment.group.name}` : "Group", value: { GroupID: enrollment.groupID } })
    }
    const targetInfo = (note: Note): TargetInfo => {
        if (note.GroupID > 0n) {
            const group = groups.find(g => g.ID === note.GroupID)
            return { icon: "fa-users", text: group?.name ?? "Group" }
        }
        return { icon: "fa-user", text: user.Name }
    }

    return (
        <div className="space-y-4">
            <Link to={`${root}/members`} className="link link-hover text-primary text-sm">← Back to members</Link>

            {/* Identity + quick stats */}
            <div className="card bg-base-100 shadow-xl">
                <div className="card-body">
                    <div className="flex items-center gap-4">
                        <Avatar src={user.AvatarURL} alt={`${user.Name}'s avatar`} size="w-16" />
                        <div className="flex-1">
                            <h2 className="text-2xl font-bold flex items-center gap-2">
                                {user.Name}
                                <Badge type="solid" color={enrollment.status} text={EnrollmentStatus[enrollment.status]} />
                            </h2>
                            <div className="text-sm text-base-content/70 flex flex-wrap gap-x-4">
                                {user.Email && <a href={`mailto:${user.Email}`} className="link link-hover">{user.Email}</a>}
                                {user.StudentID && <span>ID: {user.StudentID}</span>}
                                {enrollment.group?.name && <span>Group: {enrollment.group.name}</span>}
                            </div>
                        </div>
                        <a href={userRepoLink(user, course)} target="_blank" rel="noopener noreferrer"
                            className="btn btn-sm btn-outline gap-1">
                            <i className="fab fa-github" /> Repository
                        </a>
                    </div>

                    <div className="stats stats-horizontal shadow mt-2">
                        <Stat label="Slip-days remaining" value={enrollment.slipDaysRemaining.toString()} />
                        <Stat label="Approved" value={enrollment.totalApproved.toString()} />
                        <Stat label="Last activity" value={getFormattedTime(enrollment.lastActivityDate) || "—"} />
                    </div>
                </div>
            </div>

            {/* Submissions */}
            <div className="card bg-base-100 shadow-xl">
                <div className="card-body">
                    <h3 className="card-title text-lg">Submissions</h3>
                    <SubmissionTable assignments={assignments} submissions={submissions} userID={user.ID} root={root} />
                </div>
            </div>

            {/* Internal notes */}
            <div className="card bg-base-100 shadow-xl">
                <div className="flex items-center gap-2 bg-warning text-warning-content px-6 py-3 rounded-t-2xl">
                    <i className="fas fa-lock" />
                    <h3 className="text-md font-bold">Internal Notes</h3>
                    {notes.length > 0 && <span className="badge badge-sm">{notes.length}</span>}
                </div>
                <div className="card-body p-0">
                    <NotePanelBody notes={notes} targets={targets} targetInfo={targetInfo} />
                </div>
            </div>
        </div>
    )
}

const Stat = ({ label, value }: { label: string, value: string }) => (
    <div className="stat">
        <div className="stat-title">{label}</div>
        <div className="stat-value text-2xl">{value}</div>
    </div>
)

const SubmissionTable = ({ assignments, submissions, userID, root }: { assignments: Assignment[], submissions: Submission[], userID: bigint, root: string }) => (
    <div className="overflow-x-auto">
        <table className="table table-zebra">
            <thead>
                <tr><th>Assignment</th><th>Score</th><th>Status</th><th /></tr>
            </thead>
            <tbody>
                {assignments.length === 0 && (
                    <tr><td colSpan={4} className="text-base-content/60">No assignments.</td></tr>
                )}
                {assignments.map(assignment => {
                    const submission = submissions.find(s => s.AssignmentID === assignment.ID)
                    return (
                        <SubmissionRow key={assignment.ID.toString()}
                            name={assignment.name}
                            submission={submission}
                            userID={userID}
                            resultsLink={submission ? `${root}/results?id=${submission.ID}` : undefined}
                        />
                    )
                })}
            </tbody>
        </table>
    </div>
)

const SubmissionRow = ({ name, submission, userID, resultsLink }: { name: string, submission?: Submission, userID: bigint, resultsLink?: string }) => {
    const status = submission ? getStatusByUser(submission, userID) : Submission_Status.NONE
    const { color, icon } = submissionStatusConfig[status]
    return (
        <tr>
            <td>{name}</td>
            <td>{submission ? `${submission.score} %` : "—"}</td>
            <td className={`font-medium ${color}`}>
                {icon && <i className={`fas ${icon} mr-1`} />}
                {SubmissionStatus[status]}
            </td>
            <td>{resultsLink && <Link to={resultsLink} className="link link-hover text-primary">Open</Link>}</td>
        </tr>
    )
}

export default StudentDetails
