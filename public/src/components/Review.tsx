import React, { useEffect, useState } from "react"
import { sortByField } from "../Helpers"
import { useOvermind } from "../overmind"
import { Assignment, Enrollment, EnrollmentLink, Submission, SubmissionLink, User} from "../../proto/ag_pb"
import { useParams } from "react-router"
import Lab from "./Lab"


const Review = () => {
    const {state, actions} = useOvermind()
    const course = useParams<{id?: string}>()
    const courseID = Number(course.id)

    const [submission, setSubmission] = useState<number | undefined>(undefined)
    const [assignment, setAssignment] = useState<number | undefined>(undefined)
    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }

    }, [state.courseSubmissions, submission, setSubmission])

    const updateStatus = (status: Submission.Status, submission?: Submission) => {
        if (submission) {
            let s = new Submission()
            s.setId(submission.getId())
            s.setStatus(status)
            s.setReleased(submission.getReleased())
            s.setScore(submission.getScore())
            actions.updateSubmission({courseID: courseID, submission: s})
        }
    }

    const ReviewSubmissionsListItem = ({ submissionLink }: { submissionLink: SubmissionLink }) => {
        return (
                <li className="list-group-item" >
                    <span  onClick={() => { setSubmission(submissionLink.getSubmission()?.getId()), setAssignment(submissionLink.getAssignment()?.getId())}}>{submissionLink.getAssignment()?.getName()} - {submissionLink.getSubmission()?.getScore()} / 100</span>
                    <button style={{float: "right"}} onClick={() => updateStatus(Submission.Status.REJECTED, submissionLink.getSubmission())}>
                        Reject
                    </button>
                    <button style={{float: "right"}} onClick={() => updateStatus(Submission.Status.APPROVED, submissionLink.getSubmission())}>
                        Approve
                    </button>
                </li>
        )
    }
    

    if (state.courseSubmissions[courseID]) {
        const ReviewSubmissionsTable = sortByField(state.courseSubmissions[courseID], [EnrollmentLink.prototype.getEnrollment, Enrollment.prototype.getUser], User.prototype.getName, false).map((link: EnrollmentLink, index) => {
                return (<div key={index} className="card well" style={{width: "400px", marginBottom: "5px"}}>
                    <div key={"header"} className="card-header">
                        {link.getEnrollment()?.getUser()?.getName()} - {link.getEnrollment()?.getSlipdaysremaining()}
                    </div>
                    <ul key={"list"} className="list-group list-group-flush">
                        {link.getSubmissionsList().map((submissionLink, index) => 
                            <ReviewSubmissionsListItem key={index} submissionLink={submissionLink} />
                        )}
                    </ul>
                </div>
                )
        })
    

        return (
            <div className="review box">
                <div className="reviewTable">
                    {ReviewSubmissionsTable}
                </div>

                
                { // If submission & assignment is set by clicking an entry in ReviewSubmissionsListItem, the Lab will be displayed next to it
                submission && assignment ? (
                
                    <div className="reviewLab">
                        <Lab submissionID={submission} assignmentID={assignment} />
                    </div> )

                : "" }
                    
                
                
            </div>
        )
    }
    return (
        <div>Loading</div>
    )
}

export default Review