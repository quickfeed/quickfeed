import React, { useEffect, useState } from "react"
import { Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { generateAssignmentsHeader, generateSubmissionRows, getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import DynamicTable, { CellElement } from "./DynamicTable"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Search from "./Search"


const Results = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const [groupView, setGroupView] = useState<boolean>(false)

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => actions.setActiveSubmissionLink(undefined)
    }, [state.courseSubmissions])

    if (!state.courseSubmissions[courseID]) {
        return <h1>Fetching Submissions...</h1>
    }

    const getSubmissionCell = (submissionLink: SubmissionLink): CellElement => {
        const submission = submissionLink.getSubmission()
        if (submission) {
            return ({
                value: `${submission.getScore()} %`,
                className: submission.getStatus() === Submission.Status.APPROVED ? "result-approved" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                }
            })
        } else {
            return ({
                value: "N/A",
                onClick: () => actions.setActiveSubmissionLink(undefined)
            })
        }
    }

    const base = groupView ? ["Name"] : ["Name", "Group"]
    const header = generateAssignmentsHeader(base, state.assignments[courseID], groupView)
    const links = groupView ? state.courseGroupSubmissions[courseID] : state.courseSubmissions[courseID]
    const results = generateSubmissionRows(links, getSubmissionCell, true)

    return (
        <div>
            <div className="row">
                <div className="col">
                    <Search /><span onClick={() => setGroupView(!groupView)}>Switch View</span>
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
