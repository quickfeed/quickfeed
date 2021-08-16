import React, { useEffect, useState } from "react"
import { useActions, useAppState } from "../overmind"
import { Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import Lab from "./Lab"
import { getCourseID } from "../Helpers"
import Search from "./Search"
import { json } from "overmind"
import SubmissionApproval from "./SubmissionApproval"
import ReviewForm from "./forms/ReviewForm"


const Review = () => {
    const state = useAppState()
    const actions = useActions()

    const courseID = getCourseID()

    const [selected, setSelected] = useState<number>(0)
    const [hideApproved, setHideApproved] = useState<boolean>(false)
    const [selectedSubLink, setSelectedSubLink] = useState<SubmissionLink>()

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }

    }, [])

    const ReviewSubmissionsListItem = ({submissionLink, userIndex}: { submissionLink: SubmissionLink, userIndex: number}) => {
        const submission = json(submissionLink.getSubmission())
        const assignment = json(submissionLink.getAssignment())

        let reviews: JSX.Element | null = null

        if (assignment) { 
            reviews = assignment.getReviewers() > 0 ? <span className="float-right">{submission ? submission.getReviewsList().length : 0}/{assignment.getReviewers()}</span> : null
        }
        return (
                <li className="list-group-item" 
                    onClick={() => { actions.setActiveSubmission(json(submission)); setSelectedSubLink(submissionLink)}} 
                    hidden={selected !== assignment?.getId() && selected !== 0 || hideApproved && submission?.getStatus() == Submission.Status.APPROVED}
                >
                    <span>
                        {assignment?.getName()} - {submission?.getScore()} / 100
                    </span>
                    {reviews}
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
                    { selectedSubLink ? 
                        <ReviewForm submissionLink={selectedSubLink} setSelected={setSelectedSubLink} /> : null
                    }
                    
                    { // If submission & assignment is set by clicking an entry in ReviewSubmissionsListItem, the Lab will be displayed next to it
                    state.activeSubmission ? (
                    
                        <div className="reviewLab">
                            <SubmissionApproval />
                            <Lab teacherSubmission={state.activeSubmission} />
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