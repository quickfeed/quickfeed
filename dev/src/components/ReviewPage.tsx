import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import { Assignment, Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded } from "../Helpers"
import Search from "./Search"
import { json } from "overmind"
import ReviewForm from "./forms/ReviewForm"
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
        if (submissionLink.hasSubmission() && submissionLink.hasAssignment() && isManuallyGraded(submissionLink.getAssignment() as Assignment)) {
            return ({
                value: `${json(submissionLink.getSubmission())?.getReviewsList().length} / ${(submissionLink.getAssignment() as Assignment).getReviewers()}`,
                className: submissionLink.getSubmission()?.getStatus() === Submission.Status.APPROVED ? "result-approved" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                }
            })
        }
        else {
            return ({
                value: "N/A",
                onClick: () => { actions.setActiveSubmissionLink(submissionLink) }
            })
        }
    }

    const header = ["Name"].concat(state.assignments[courseID].map(assignment => assignment.getName()))


    const data = state.courseSubmissionsList[courseID]?.map((link) => {
        const row: Row = []
        row.push(link.user ? { value: link.user.getName(), link: `https://github.com/${link.user.getLogin()}` } : "")
        if (link.submissions && link.user) {
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
