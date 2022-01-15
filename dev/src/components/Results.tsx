import React, { useEffect } from "react"
import { Group, Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import DynamicTable, { CellElement, Row } from "./DynamicTable"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Search from "./Search"


const Results = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => actions.setActiveSubmissionLink(undefined)
    }, [state.courseSubmissions])

    if (!state.courseSubmissions[courseID]) {
        return <h1>Fetching Submissions...</h1>
    }

    const header = ["Name", "Group"].concat(state.assignments[courseID].map(assignment => {
        return assignment.getName()
    }))


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

    const results = state.courseSubmissions[courseID].map(link => {
        const data: Row = []
        if (link.user && link.enrollment) {
            data.push({ value: link.user.getName(), link: `https://github.com/${link.user.getLogin()}` })
            data.push(link.enrollment.hasGroup() ? (link.enrollment.getGroup() as Group)?.getName() : "")
            if (link.submissions) {
                for (const submissionLink of link.submissions) {
                    data.push(getSubmissionCell(submissionLink))
                }
            }
        }
        return data
    })

    return (
        <div>
            <div className="row">
                <div className="col">
                    <Search />
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
