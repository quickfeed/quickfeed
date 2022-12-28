import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { Assignment, Submission } from '../../proto/qf/types_pb'
import { hasReviews, isManuallyGraded } from '../Helpers'
import { useAppState, useActions } from '../overmind'
import CourseUtilityLinks from './CourseUtilityLinks'
import LabResultTable from './LabResultTable'
import ReviewResult from './ReviewResult'

interface MatchProps {
    id: string
    lab: string
}

/** Lab displays a submission based on the /course/:id/lab/:lab route if the user is a student.
 *  If the user is a teacher, Lab displays a submission based on the submission in state.currentSubmission.
 */
const Lab = (): JSX.Element => {

    const state = useAppState()
    const actions = useActions()
    const { id, lab } = useParams<MatchProps>()
    const courseID = id
    const assignmentID = lab ? BigInt(lab) : BigInt(-1)

    useEffect(() => {
        if (!state.isTeacher) {
            actions.setActiveAssignment(Number(lab))
        }
    }, [lab])

    const Lab = () => {
        let submission: Submission | null
        let assignment: Assignment | null

        if (state.isTeacher) {
            // If used for grading purposes, retrieve submission from state.currentSubmission
            submission = state.currentSubmission
            assignment = state.assignments[courseID].find(a => a.ID === submission?.AssignmentID) ?? null
        } else {
            // Retrieve the student's submission
            submission = state.submissions[courseID]?.find(s => s.AssignmentID === assignmentID) ?? null
            assignment = state.assignments[courseID]?.find(a => a.ID === assignmentID) ?? null
        }

        if (assignment && submission) {
            // Confirm both assignment and submission exists before attempting to render
            const review = hasReviews(submission) ? submission.reviews : []
            let buildLog: JSX.Element[] = []
            const buildLogRaw = submission.BuildInfo?.BuildLog
            if (buildLogRaw) {
                buildLog = buildLogRaw.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>)
            }

            return (
                <div key={submission.ID.toString()} className="mb-4">
                    <LabResultTable submission={submission} assignment={assignment} />

                    {isManuallyGraded(assignment) && submission.released ? <ReviewResult review={review[0]} /> : null}

                    <div className="card bg-light">
                        <code className="card-body" style={{ color: "#c7254e" }}>{buildLog}</code>
                    </div>
                </div>
            )
        }
        return (
            <div>No submission found</div>
        )
    }

    return (
        <div className={state.isTeacher ? "" : "row"}>
            <div className={state.isTeacher ? "" : "col-md-9"}>
                <Lab />
            </div>
            {state.isTeacher ? null : <CourseUtilityLinks />}
        </div>
    )
}

export default Lab
