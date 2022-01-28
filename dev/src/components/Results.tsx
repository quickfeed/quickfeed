import React, { useEffect, useState } from "react"
import { useHistory } from "react-router"
import { Enrollment, SubmissionLink } from "../../proto/ag/ag_pb"
import { Color, generateAssignmentsHeader, generateSubmissionRows, getCourseID, isApproved, isRevision, sortCourseSubmissions, SubmissionSort } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Button, { ButtonType } from "./admin/Button"
import DynamicTable, { CellElement } from "./DynamicTable"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Search from "./Search"


const Results = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const history = useHistory()
    const [groupView, setGroupView] = useState<boolean>(false)

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            // Redirect to home if the call to get all submissions fails
            actions.getAllCourseSubmissions(courseID).then((success => {
                if (!success) {
                    history.push("/")
                }
            }))
        }
        return () => {
            actions.setActiveSubmissionLink(undefined)
            actions.setActiveEnrollment(undefined)
        }
    }, [])

    if (!state.courseSubmissions[courseID]) {
        return <h1>Fetching Submissions...</h1>
    }

    const getSubmissionCell = (submissionLink: SubmissionLink, enrollment: Enrollment): CellElement => {
        const submission = submissionLink.getSubmission()
        if (submission) {
            return ({
                value: `${submission.getScore()} %`,
                className: isApproved(submission) ? "result-approved" : isRevision(submission) ? "result-revision" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                    actions.setActiveEnrollment(enrollment)
                }
            })
        } else {
            return ({
                value: "N/A",
                onClick: () => {
                    actions.setActiveSubmissionLink(undefined)
                    actions.setActiveEnrollment(undefined)
                }
            })
        }
    }

    const base = groupView ? ["Name"] : ["Name", "Group"]
    const assignments = state.assignments[courseID].filter(assignment => (state.review.assignmentID < 0) || assignment.getId() === state.review.assignmentID)
    const header = generateAssignmentsHeader(base, assignments, groupView)
    const links = groupView ? state.courseGroupSubmissions[courseID] : state.courseSubmissions[courseID]
    const results = generateSubmissionRows(links, getSubmissionCell, true)

    return (
        <div>
            <div className="row">
                <div className={state.review.assignmentID >= 0 ? "col-md-4" : "col-md-6"}>
                    <Search >
                        <Button type={ButtonType.BUTTON}
                            text={groupView ? "View by group" : "View by student"}
                            onclick={() => { setGroupView(!groupView); actions.review.setAssignmentID(-1) }}
                            color={groupView ? Color.BLUE : Color.GREEN} />
                    </Search>
                    <DynamicTable header={header} data={results} />
                </div>
                <div className="col reviewLab">
                    {state.currentSubmission ?
                        <>
                            <ManageSubmissionStatus />
                            <div className="reviewLabResult mt-2">
                                <Lab />
                            </div>
                        </>
                        : null}
                </div>
            </div>
        </div>

    )
}

export default Results
