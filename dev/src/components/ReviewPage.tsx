import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import { Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded } from "../Helpers"
import Search from "./Search"
import { json } from "overmind"
import ReviewForm from "./manual-grading/ReviewForm"
import DynamicTable, { CellElement, Row } from "./DynamicTable"

const ReviewPage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => { actions.setActiveSubmissionLink(undefined) }
    }, [])

    const generateReviewCell = (submissionLink: SubmissionLink): CellElement => {
        const submission = submissionLink.getSubmission()
        const assignment = submissionLink.getAssignment()
        if (submission && assignment && isManuallyGraded(assignment)) {
            return ({
                value: `${json(submission).getReviewsList().length} / ${assignment.getReviewers()}`,
                className: submission.getStatus() === Submission.Status.APPROVED ? "result-approved" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                }
            })
        } else {
            return ({
                value: "N/A",
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                }
            })
        }
    }

    const header = ["Name"].concat(state.assignments[courseID].map(assignment => assignment.getName()))


    const data = state.courseSubmissions[courseID]?.map((link) => {
        const row: Row = []
        if (link.submissions && link.user) {
            row.push({ value: link.user.getName(), link: `https://github.com/${link.user.getLogin()}` })
            link.submissions.forEach(submission => {
                row.push(generateReviewCell(submission))
            })
        }
        return row
    })

    return (
        <div>
            <div className="row">
                <div className="col-md-6">
                    <Search placeholder={"Search by name ..."} />
                    <DynamicTable header={header} data={data} />
                </div>
                {state.activeSubmissionLink ? <ReviewForm /> : null}
            </div>
        </div>
    )
}

export default ReviewPage
