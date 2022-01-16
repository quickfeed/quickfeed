import React, { useEffect, useState } from "react"
import { useActions, useAppState } from "../overmind"
import { Submission, SubmissionLink } from "../../proto/ag/ag_pb"
import { Color, generateAssignmentsHeader, generateSubmissionRows, getCourseID, isManuallyGraded } from "../Helpers"
import Search from "./Search"
import { json } from "overmind"
import ReviewForm from "./manual-grading/ReviewForm"
import DynamicTable, { RowElement } from "./DynamicTable"
import Button, { ButtonType } from "./admin/Button"


const ReviewPage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const [groupView, setGroupView] = useState<boolean>(false)

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => { actions.setActiveSubmissionLink(undefined) }
    }, [])

    const generateReviewCell = (submissionLink: SubmissionLink): RowElement => {
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

    const links = groupView ? state.courseGroupSubmissions[courseID] : state.courseSubmissions[courseID]
    const rows = generateSubmissionRows(links, generateReviewCell)
    const header = generateAssignmentsHeader(["Name"], state.assignments[courseID], groupView)

    return (
        <div>
            <div className="row">
                <div className="col-md-6">
                    <div>
                        <Search placeholder={"Search by name ..."} />
                        <Button type={ButtonType.BUTTON}
                            text={groupView ? "View by group" : "View by student"}
                            onclick={() => setGroupView(!groupView)}
                            color={groupView ? Color.BLUE : Color.GREEN} />
                    </div>
                    <DynamicTable header={header} data={rows} />
                </div>
                {state.activeSubmissionLink ? <ReviewForm /> : null}
            </div>
        </div>
    )
}

export default ReviewPage
