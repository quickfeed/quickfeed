import React, { useEffect, useState } from "react"
import { useActions, useAppState } from "../overmind"
import { SubmissionLink } from "../../proto/ag/ag_pb"
import Lab from "./Lab"
import { getCourseID, isManuallyGraded } from "../Helpers"
import Search from "./Search"
import { json } from "overmind"
import ManageSubmissionStatus from "./ManageSubmissionStatus"

// TODO: Refactor & come up with better visuals
const Review = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    const courseID = getCourseID()

    const [selected, setSelected] = useState<number>(0)
    const [hideNonManual, setHideNonManual] = useState<boolean>(false)
    const [selectedSubLink, setSelectedSubLink] = useState<SubmissionLink>()

    useEffect(() => {
        if (courseID && !state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }

    }, [])

    const ReviewSubmissionsListItem = ({submissionLink}: { submissionLink: SubmissionLink }) => {
        const submission = json(submissionLink.getSubmission())
        const assignment = json(submissionLink.getAssignment())

        let reviews: JSX.Element | null = null
        let className = "list-group-item"
        if (assignment) { 
            reviews = isManuallyGraded(assignment) ? <span className="float-right">{submission ? submission.getReviewsList().length : 0}/{assignment.getReviewers()}</span> : null
            className = !isManuallyGraded(assignment) ? className + " list-group-item-secondary" : className
        }
        return (
                <li className={className}
                    onClick={() => { actions.setActiveSubmission(json(submission)); setSelectedSubLink(submissionLink)}} 
                    hidden={selected !== assignment?.getId() && selected !== 0 || hideNonManual && (assignment && !isManuallyGraded(assignment))}
                >
                    <span>
                        {assignment?.getName()} - {submission?.getScore()} / 100
                    </span>
                    {reviews}
                </li>
        )
    }
    

    if (state.courseSubmissions[courseID]) {
        const ReviewSubmissionsTable = state.courseSubmissions[courseID].map(user => {
            if (user.enrollment && user.submissions) {
                
                return (
                    <div key={user.enrollment.getId()} className="card well" style={{marginBottom: "5px"}} hidden={!user.user?.getName().toLowerCase().includes(state.query)}>
                        <div key={"header"} className="card-header">
                            {user.user?.getName()}
                        </div>
                        <ul key={"list"} className="list-group list-group-flush">
                            {user.submissions.map((submissionLink, index) => 
                                <ReviewSubmissionsListItem key={index} submissionLink={submissionLink} />
                            )}
                        </ul>
                    </div>
                )
            }
        })
    
        const Options = state.assignments[courseID].map(assignment => {
            return <option key={`assignment-${assignment.getId()}`} value={assignment.getId()}>{assignment.getName()}</option>
        })

        return (
            <div className="box">
                <select onChange={e => setSelected(Number(e.currentTarget.value))}>
                    <option value={0}>All Submissions</option>
                    {Options}
                </select>
                <input type={"checkbox"} checked={hideNonManual} onChange={(e) => setHideNonManual(e.target.checked)}></input>

                <button onClick={() => actions.getAllCourseSubmissions(courseID)}>Refresh ... </button>
                <div className="row">
                    
                    <div className="col-md-6">
                        <Search placeholder={"Search by name ..."} />
                        {ReviewSubmissionsTable}
                    </div>
                    { //selectedSubLink ? 
                      //  <ReviewForm submissionLink={selectedSubLink} setSelected={setSelectedSubLink} /> : null
                    }
                    
                    { // If submission & assignment is set by clicking an entry in ReviewSubmissionsListItem, the Lab will be displayed next to it
                    state.activeSubmission ? (
                    
                        <div className="reviewLab col">
                            <ManageSubmissionStatus />
                            <div className="reviewLabResult">
                            <Lab teacherSubmission={state.activeSubmission} />
                            </div>
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