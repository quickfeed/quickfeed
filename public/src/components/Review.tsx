import React, { useEffect, useState } from "react"
import { useOvermind } from "../overmind"
import { Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import Lab from "./Lab"
import { getCourseID } from "../Helpers"
import Search from "./Search"


const Review = () => {
    const {state, actions} = useOvermind()
    const courseID = getCourseID()

    const [submission, setSubmission] = useState<number | undefined>(undefined)
    const [assignment, setAssignment] = useState<number | undefined>(undefined)
    const [selected, setSelected] = useState<number>(0)
    const [hideApproved, setHideApproved] = useState<boolean>(false)

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }

    }, [])

    const updateStatus = (status: Submission.Status, submission?: Submission, userIndex?: number, submissionIndex?: number) => {
        if (submission && userIndex && submissionIndex) {
            let s = new Submission()
            s.setId(submission.getId())
            s.setStatus(status)
            s.setReleased(submission.getReleased())
            s.setScore(submission.getScore())
            actions.updateSubmission({courseID: courseID, submission: s, userIndex: userIndex, submissionIndex: submissionIndex - 1})
        }
    }

    const ReviewSubmissionsListItem = (props: { submissionLink: SubmissionLink, userIndex: number}) => {
        return (
                <li className="list-group-item" hidden={selected !== props.submissionLink.getAssignment()?.getId() && selected !== 0 || hideApproved && props.submissionLink.getSubmission()?.getStatus() == Submission.Status.APPROVED}>
                    <span  onClick={() => { setSubmission(props.submissionLink.getSubmission()?.getId()), setAssignment(props.submissionLink.getAssignment()?.getId())}}>{props.submissionLink.getAssignment()?.getName()} - {props.submissionLink.getSubmission()?.getScore()} / 100</span>
                    <button style={{float: "right"}} onClick={() => {updateStatus(Submission.Status.REJECTED, props.submissionLink.getSubmission(), props.userIndex, props.submissionLink.getAssignment()?.getOrder())}}>
                        Reject
                    </button>
                    <button style={{float: "right"}} onClick={() => updateStatus(Submission.Status.APPROVED, props.submissionLink.getSubmission(), props.userIndex, props.submissionLink.getAssignment()?.getOrder())}>
                        Approve
                    </button>
                </li>
        )
    }
    

    if (state.courseSubmissions[courseID]) {
        const ReviewSubmissionsTable = state.courseSubmissions[courseID].map((user, userIndex) => {
            if (user.enrollment && user.submissions) {
                return (
                    <div className="card well" style={{width: "400px", marginBottom: "5px"}} hidden={!user.user?.getName().toLowerCase().includes(state.query)}>
                        <div key={"header"} className="card-header">
                            {user.user?.getName()}
                        </div>
                        <ul key={"list"} className="list-group list-group-flush">
                            {user.submissions.map((submissionLink, index) => 
                                <ReviewSubmissionsListItem key={index} submissionLink={submissionLink} userIndex={userIndex} />
                            )}
                        </ul>
                    </div>
                )
            }
        })
    
        const Options = state.assignments[courseID].map(assignment => {
            return <option value={assignment.getId()}>{assignment.getName()}</option>
        })

        return (
            <div className="box">
                <select onChange={e => setSelected(Number(e.currentTarget.value))}>
                    <option value={0}>All Submissions</option>
                    {Options}
                </select>
                <input type={"checkbox"} checked={hideApproved} onChange={(e) => setHideApproved(e.target.checked)}></input>
                <Search placeholder={"Search by name ..."} />
                <button onClick={() => actions.getAllCourseSubmissions(courseID)}>Refresh ... </button>
                <div className="review">
                    
                    <div className="reviewTable">
                        {ReviewSubmissionsTable}
                    </div>

                    
                    { // If submission & assignment is set by clicking an entry in ReviewSubmissionsListItem, the Lab will be displayed next to it
                    submission && assignment ? (
                    
                        <div className="reviewLab">
                            <Lab submissionID={submission} assignmentID={assignment} />
                        </div> )

                    : null }
                        
                    
                    
                </div>
            </div>
        )
    }
    return (
        <div>Loading</div>
    )
}

export default Review