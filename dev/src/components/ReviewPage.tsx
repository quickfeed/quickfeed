import React, { useEffect, useState } from "react"
import { useActions, useAppState } from "../overmind"
import { SubmissionLink } from "../../proto/ag/ag_pb"
import { Color, generateAssignmentsHeader, generateSubmissionRows, getCourseID, isApproved, isManuallyGraded, isPending, isRevision } from "../Helpers"
import Search from "./Search"
import ReviewForm from "./manual-grading/ReviewForm"
import DynamicTable, { RowElement } from "./DynamicTable"
import Button, { ButtonType } from "./admin/Button"
import Release from "./Release"


const ReviewPage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const isCourseCreator = state.courses[courseID].getCoursecreatorid() === state.self.getId()
    const [groupView, setGroupView] = useState<boolean>(false)

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => {
            actions.setActiveSubmissionLink(undefined)
            actions.review.setAssignmentID(-1)
        }
    }, [])

    const generateReviewCell = (submissionLink: SubmissionLink): RowElement => {
        const submission = submissionLink.getSubmission()
        const assignment = submissionLink.getAssignment()
        if (submission && assignment && isManuallyGraded(assignment)) {
            const reviews = state.review.reviews[courseID][submission.getId()] ?? []
            const isSelected = state.activeSubmission === submission?.getId()
            const score = reviews.reduce((acc, review) => acc + review.getScore(), 0) / reviews.length
            const willBeReleased = state.review.minimumScore > 0 && score >= state.review.minimumScore
            return ({
                // TODO: Figure out a better way to visualize released submissions than '(r)'
                value: `${reviews.length}/${assignment.getReviewers()} ${submission.getReleased() ? "(r)" : ""}`,
                className: `${isApproved(submission) ? "result-approved" : isRevision(submission) ? "result-revision" : "result-pending"} ${isSelected ? "selected" : ""} ${willBeReleased ? "release" : ""}`,
                onClick: () => {
                    actions.setActiveSubmissionLink(submissionLink)
                    actions.review.setSelectedReview(-1)
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
    const assignments = state.assignments[courseID].filter(assignment => (state.review.assignmentID < 0) || assignment.getId() === state.review.assignmentID)
    const header = generateAssignmentsHeader(["Name"], assignments, groupView)

    return (
        <div>
            {isCourseCreator && <Release />}
            <div className="row">
                <div className={state.review.assignmentID >= 0 ? "col-md-4" : "col-md-6"}>
                    <Search placeholder={"Search by name ..."} >
                        <Button type={ButtonType.BUTTON}
                            text={groupView ? "View by group" : "View by student"}
                            onclick={() => { setGroupView(!groupView); actions.review.setAssignmentID(-1) }}
                            color={groupView ? Color.BLUE : Color.GREEN} />
                    </Search>
                    <DynamicTable header={header} data={rows} />
                </div>
                {state.activeSubmissionLink ? <ReviewForm /> : null}
            </div>
        </div>
    )
}

export default ReviewPage
