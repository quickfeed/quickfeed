import { json } from 'overmind'
import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { Assignment, Submission } from '../../proto/ag/ag_pb'
import { BuildInfo } from '../../proto/kit/score/score_pb'
import { isManuallyGraded } from '../Helpers'
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
const Lab = ({teacherSubmission}: {teacherSubmission?: Submission}): JSX.Element => {
    
    const state = useAppState()
    const actions = useActions()
    const {id ,lab} = useParams<MatchProps>()
    const courseID = Number(id)
    const assignmentID = Number(lab)

    useEffect(() => {
        // Do not start the commit hash fetch-loop for submissions that are not personal
        if (!teacherSubmission) {
            actions.setActiveLab(assignmentID)

            // TODO: Implement SubmissionCommitHash
            /*
            const ping = setInterval(() => {  
                actions.getSubmissionCommitHash({courseID: courseID, assignmentID: assignmentID})
            }, 5000)

            return () => {clearInterval(ping), actions.setActiveLab(-1)}
            */
        }
    }, [lab])

    
    const Lab = () => {
        let submission: Submission | undefined
        let assignment: Assignment | undefined

        // If used for grading purposes, retrieve submission from courseSubmissions
        if (teacherSubmission) {
            submission = teacherSubmission
            assignment = state.assignments[courseID].find(a => a.getId() === submission?.getAssignmentid())
        } 
        // Retreive personal submission
        else {
            submission = state.submissions[courseID]?.find(s => s.getAssignmentid() === assignmentID)
            assignment = state.assignments[courseID]?.find(a => a.getId() === assignmentID)
        }
        
        // Confirm both assignment and submission exists before attempting to render
        if (assignment && submission) {
            const review = json(submission).getReviewsList()
            let buildLogElement: JSX.Element[] = []
            
            const buildLog = submission.hasBuildinfo() ? (submission.getBuildinfo() as BuildInfo).getBuildlog() : null
            if (buildLog){
                buildLogElement = buildLog.split("\\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>);
            }

            return (
                <div key={submission.getId()}>

                    <LabResultTable submission={submission} assignment={assignment} />

                    {isManuallyGraded(assignment) ? <ReviewResult review={review}/> : null}

                    <div className="card bg-light">
                        <code className="card-body" style={{color: "#c7254e"}}>{buildLogElement}</code>
                    </div>
                </div>
            )
        }
        return (
            <div>No submission found</div>
        )
    }

    return (
        <div className={teacherSubmission ? "" : "row"}>
            <div className={teacherSubmission ? "" : "col-md-9"}>
                <Lab />
            </div>
            {teacherSubmission ? null : <CourseUtilityLinks courseID={courseID} />}
        </div>
    )
}

export default Lab