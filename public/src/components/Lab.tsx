import React, { useEffect } from 'react'
import { useLocation, useParams } from 'react-router'
import { Assignment, Submission } from '../../proto/qf/types_pb'
import { hasReviews, isManuallyGraded } from '../Helpers'
import { useAppState, useActions } from '../overmind'
import CourseLinks from "./CourseLinks"
import LabResultTable from "./LabResultTable"
import ReviewResult from './ReviewResult'
import { CenteredMessage, KnownMessage } from './CenteredMessage'

interface MatchProps {
    id: string
    lab: string
}

/** Lab displays a submission based on the /course/:id/lab/:lab route if the user is a student.
 *  If the user is a teacher, Lab displays the currently selected submission.
 */
const Lab = () => {
    const state = useAppState()
    const actions = useActions().global
    const { id, lab } = useParams<MatchProps>()
    const courseID = id
    const assignmentID = lab ? BigInt(lab) : BigInt(-1)
    const location = useLocation()
    const isGroupLab = location.pathname.includes("group-lab")

    useEffect(() => {
        if (!state.isTeacher) {
            actions.setSelectedAssignmentID(Number(lab))
        }
    }, [lab])

    const InternalLab = () => {
        let submission: Submission | null
        let assignment: Assignment | null

        if (state.isTeacher) {
            // If used for grading purposes, retrieve the currently selected submission
            submission = state.selectedSubmission
            assignment = state.assignments[courseID].find(a => a.ID === submission?.AssignmentID) ?? null
        } else {
            // Retrieve the student's submission
            assignment = state.assignments[courseID]?.find(a => a.ID === assignmentID) ?? null
            if (!assignment) {
                return <CenteredMessage message={KnownMessage.NoAssignment} />
            }
            const submissions = state.submissions.ForAssignment(assignment) ?? null
            if (!submissions) {
                return <CenteredMessage message={KnownMessage.NoSubmission} />
            }

            if (isGroupLab) {
                submission = submissions.find(s => s.groupID > 0n) ?? null
            } else {
                submission = submissions.find(s => s.userID === state.self.ID && s.groupID === 0n) ?? null
            }
        }

        if (assignment && submission) {
            // Confirm both assignment and submission exists before attempting to render
            const review = hasReviews(submission) ? submission.reviews : []
            let buildLog: React.JSX.Element[] = []
            const buildLogRaw = submission.BuildInfo?.BuildLog
            if (buildLogRaw) {
                buildLog = buildLogRaw.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>)
            }

            return (
                <div key={submission.ID.toString()} className="mb-4">
                    <LabResultTable submission={submission} assignment={assignment} />

                    {isManuallyGraded(assignment.reviewers) && submission.released ? <ReviewResult review={review[0]} /> : null}

                    <div className="card bg-light">
                        <code className="card-body" style={{ color: "#c7254e", wordBreak: "break-word" }}>{buildLog}</code>
                    </div>
                </div>
            )
        }
        return <CenteredMessage message={KnownMessage.NoSubmission} />
    }

    return (
        <div className={state.isTeacher ? "" : "row"}>
            <div className={state.isTeacher ? "" : "col-md-9"}>
                <InternalLab />
            </div>
            {state.isTeacher ? null : <CourseLinks />}
        </div>
    )
}

export default Lab
