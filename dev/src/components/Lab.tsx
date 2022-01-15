import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { Assignment, Submission } from '../../proto/ag/ag_pb'
import { hasReviews, isManuallyGraded } from '../Helpers'
import { useAppState, useActions } from '../overmind'
import CourseUtilityLinks from './CourseUtilityLinks'
import LabResultTable from './LabResultTable'
import ReviewResult from './ReviewResult'

interface MatchProps {
    id: string
    lab: string
}

/** Lab displays a submission based on the /course/:id/:lab route if the user is a student.
 *  If the user is a teacher, Lab displays a submission based on the submission in state.currentSubmission.
 */
const Lab = (): JSX.Element => {

    const state = useAppState()
    const actions = useActions()
    const { id, lab } = useParams<MatchProps>()
    const courseID = Number(id)
    const assignmentID = Number(lab)

    useEffect(() => {
        if (!state.isTeacher) {
            actions.setActiveAssignment(assignmentID)
        }
    }, [lab])

    const Lab = () => {
        let submission: Submission | undefined
        let assignment: Assignment | undefined

        if (state.isTeacher) {
            // If used for grading purposes, retrieve submission from state.currentSubmission
            submission = state.currentSubmission
            assignment = state.assignments[courseID].find(a => a.getId() === submission?.getAssignmentid())
        } else {
            // Retrieve the student's submission
            submission = state.submissions[courseID]?.find(s => s.getAssignmentid() === assignmentID)
            assignment = state.assignments[courseID]?.find(a => a.getId() === assignmentID)
        }

        if (assignment && submission) {
            // Confirm both assignment and submission exists before attempting to render
            const review = hasReviews(submission) ? submission.getReviewsList() : []
            let buildLog: JSX.Element[] = []
            const buildLogRaw = submission.getBuildinfo()?.getBuildlog()
            if (buildLogRaw) {
                buildLog = buildLogRaw.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>)
            }

            return (
                <div key={submission.getId()}>
                    <LabResultTable submission={submission} assignment={assignment} />

                    {isManuallyGraded(assignment) ? <ReviewResult review={review[0]} /> : null}

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
