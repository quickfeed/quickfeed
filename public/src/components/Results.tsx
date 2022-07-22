import React, { useEffect } from "react"
import { Enrollment, SubmissionLink } from "../../proto/qf/types_pb"
import { Color, getCourseID, getSubmissionCellColor, isManuallyGraded, SubmissionSort } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Button, { ButtonType } from "./admin/Button"
import { generateAssignmentsHeader, generateSubmissionRows } from "./ComponentsHelpers"
import DynamicTable, { CellElement, Row, RowElement } from "./DynamicTable"
import TableSort from "./forms/TableSort"
import LabResult from "./LabResult"
import ReviewForm from "./manual-grading/ReviewForm"
import Search from "./Search"


const Results = ({ review }: { review: boolean }): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!state.courseSubmissions[courseID]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => {
            actions.setActiveSubmissionLink(undefined)
            actions.setGroupView(false)
            actions.review.setAssignmentID(-1)
            actions.setActiveEnrollment(undefined)
        }
    }, [state.courseSubmissions])

    if (!state.courseSubmissions[courseID]) {
        return <h1>Fetching Submissions...</h1>
    }


    const generateReviewCell = (submissionLink: SubmissionLink.AsObject): RowElement => {
        const submission = submissionLink.submission
        const assignment = submissionLink.assignment
        if (submission && assignment && isManuallyGraded(assignment)) {
            const reviews = state.review.reviews[assignment.courseid][submission.id] ?? []
            const isSelected = state.activeSubmission === submission.id
            const score = reviews.reduce((acc, review) => acc + review.score, 0) / reviews.length
            const willBeReleased = state.review.minimumScore > 0 && score >= state.review.minimumScore
            return ({
                // TODO: Figure out a better way to visualize released submissions than '(r)'
                value: `${reviews.length}/${assignment.reviewers} ${submission.released ? "(r)" : ""}`,
                className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""} ${willBeReleased ? "release" : ""}`,
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

    const getSubmissionCell = (submissionLink: SubmissionLink.AsObject, enrollment: Enrollment.AsObject): CellElement => {
        const submission = submissionLink.submission
        if (submission) {
            const isSelected = state.activeSubmission === submission.id
            return ({
                value: `${submission.score} %`,
                className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""}`,
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


    const groupView = state.groupView
    const base: Row = [{ value: "Name", onClick: () => actions.setSubmissionSort(SubmissionSort.Name) }]
    const assignments = state.assignments[courseID].filter(assignment => (state.review.assignmentID < 0) || assignment.id === state.review.assignmentID)
    const assignmentIDs = assignments.filter(assignment => groupView ? assignment.isgrouplab : true).map(assignment => assignment.id)
    const header = generateAssignmentsHeader(base, assignments, groupView)

    const links = state.sortedAndFilteredSubmissions
    const generator = review ? generateReviewCell : getSubmissionCell
    const rows = generateSubmissionRows(links, review, generator, assignmentIDs, false)


    return (
        <div className="row">
            <div className={state.review.assignmentID >= 0 ? "col-md-4" : "col-xl-6"}>
                <Search placeholder={"Search by name ..."} className="mb-2" >
                    <Button type={ButtonType.BUTTON}
                        classname="ml-2"
                        text={`View by ${groupView ? "student" : "group"}`}
                        onclick={() => { actions.setGroupView(!groupView); actions.review.setAssignmentID(-1) }}
                        color={groupView ? Color.BLUE : Color.GREEN} />
                </Search>
                <TableSort />
                <DynamicTable header={header} data={rows} />
            </div>
            <div className="col reviewLab">
                {review ? <ReviewForm /> : <LabResult />}
            </div>
        </div>
    )
}

export default Results
