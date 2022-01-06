import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import { Assignment, Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { getCourseID, isManuallyGraded } from "../Helpers"
import Search from "./Search"
import { json } from "overmind"
import ReviewForm from "./forms/ReviewForm"
import DynamicTable, { CellElement } from "./DynamicTable"

const ReviewPage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
    })

    const generateReviewCell = (submissionLink: SubmissionLink, enrollmentIndex: number): CellElement => {
        if (submissionLink.hasSubmission() && submissionLink.hasAssignment() && isManuallyGraded(submissionLink.getAssignment() as Assignment)) {
            return ({
                value: `${json(submissionLink.getSubmission())?.getReviewsList().length} / ${(submissionLink.getAssignment() as Assignment).getReviewers()}`,
                className: submissionLink.getSubmission()?.getStatus() === Submission.Status.APPROVED ? "result-approved" : "result-pending",
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink),
                        actions.setActiveSubmission(submissionLink.getSubmission()?.getId())
                    actions.setSelectedEnrollment(enrollmentIndex)
                }
            })
        }
        else {
            return ({
                value: "N/A",
                onClick: () => { actions.setActiveSubmission(undefined), actions.setActiveSubmissionLink(submissionLink) }
            })
        }
    }

    const header = ["Name"].concat(state.assignments[courseID].map(assignment => assignment.getName()))


    const data = state.courseSubmissionsList[courseID]?.map((link, index) => {
        const row: (string | JSX.Element | CellElement)[] = []
        row.push(link.user ? { value: link.user.getName(), link: `https://github.com/${link.user.getLogin()}` } : "")
        if (link.submissions && link.user) {
            link.submissions.forEach(submission => {
                row.push(generateReviewCell(submission, index))
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