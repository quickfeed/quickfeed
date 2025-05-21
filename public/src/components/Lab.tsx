import React, { useEffect } from 'react'
import { useLocation, useParams } from 'react-router'
import { Assignment, Submission } from '../../proto/qf/types_pb'
import { hasReviews, isManuallyGraded } from '../Helpers'
import { useAppState, useActions } from '../overmind'
import CourseLinks from "./CourseLinks"
import LabResultTable from "./LabResultTable"
import ReviewResult from './ReviewResult'
import { CenteredMessage, KnownMessage } from './CenteredMessage'


/** Lab displays a submission based on the /course/:id/lab/:lab route if the user is a student.
 *  If the user is a teacher, Lab displays the currently selected submission.
 */
const Lab = () => {
    const state = useAppState()
    const actions = useActions()
    const { id, lab } = useParams()
    const courseID = id ?? ""
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
                return <CenteredMessage message={KnownMessage.StudentNoAssignment} />
            }
            const submissions = state.submissions.ForAssignment(assignment)
            if (submissions.length === 0) {
                return <CenteredMessage message={KnownMessage.StudentNoSubmission} />
            }

            const query = (s: Submission) => isGroupLab
                ? s.groupID > 0n
                : s.userID === state.self.ID && s.groupID === 0n

            submission = submissions.find(s => query(s)) ?? null
        }

        if (assignment && submission) {
            // Confirm both assignment and submission exists before attempting to render
            const review = hasReviews(submission) ? submission.reviews : []
            let buildLog: React.JSX.Element[] = []
            const buildLogRaw = submission.BuildInfo?.BuildLog
            if (buildLogRaw) {
                // using the index as the key is not ideal, but in this case it is acceptable
                // because the log lines are not expected to change unless a new submission is made
                // in which case the component will be re-rendered anyways
                buildLog = buildLogRaw.split("\n").map((logLine: string, idx: number) => <span key={idx}>{logLine}<br /></span>) // skipcq: JS-0437
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
        return <CenteredMessage message={state.isTeacher ? KnownMessage.TeacherNoSubmission : KnownMessage.StudentNoSubmission} />
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
