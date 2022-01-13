import { json } from 'overmind'
import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { Assignment, Submission } from '../../proto/ag/ag_pb'
import { BuildInfo } from '../../proto/kit/score/score_pb'
import { hasReviews, isManuallyGraded } from '../Helpers'
import { useAppState, useActions } from '../overmind'
import CourseUtilityLinks from './CourseUtilityLinks'
import LabResultTable from './LabResultTable'
import ReviewResult from './ReviewResult'

interface MatchProps {
    id: string
    lab: string
}



/** Displays a Lab submission based on the /course/:id/:lab route
 *
 *  If used to display a lab for grading purposes, pass in a TeacherLab object
 */
const Lab = (): JSX.Element => {

    const state = useAppState()
    const actions = useActions()
    const { id, lab } = useParams<MatchProps>()
    const courseID = Number(id)
    const assignmentID = Number(lab)

    useEffect(() => {
        if (!state.isTeacher) {
            actions.setActiveLab(assignmentID)
        }
    }, [lab])


    const Lab = () => {
        let submission: Submission | undefined
        let assignment: Assignment | undefined

        // If used for grading purposes, retrieve submission from courseSubmissions
        if (state.isTeacher) {
            submission = state.currentSubmission
            assignment = state.assignments[courseID].find(a => a.getId() === submission?.getAssignmentid())
        }
        // Retreive personal submission
        else {
            submission = state.submissions[courseID]?.find(s => s.getAssignmentid() === assignmentID)
            assignment = state.assignments[courseID]?.find(a => a.getId() === assignmentID)
        }

        // Confirm both assignment and submission exists before attempting to render
        if (assignment && submission) {
            const review = hasReviews(submission) ? submission.getReviewsList() : []
            let buildLog: JSX.Element[] = []
            const buildLogRaw = submission.hasBuildinfo() ? (submission.getBuildinfo() as BuildInfo).getBuildlog() : null
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
