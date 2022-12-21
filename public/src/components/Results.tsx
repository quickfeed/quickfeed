import React, { useEffect, useMemo } from "react"
import { Enrollment, Group, Submission } from "../../proto/qf/types_pb"
import { Color, getCourseID, getSubmissionCellColor, SubmissionSort } from "../Helpers"
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

    const members = useMemo(() => { return state.courseMembers }, [state.courseMembers, state.groupView])
    const assignments = useMemo(() => { 
        // Filter out all assignments that are not the selected assignment, if any assignment is selected
        return state.assignments[courseID.toString()].filter(a => state.review.assignmentID <= 0 || a.ID === state.review.assignmentID)
    }, [state.assignments, courseID, state.review.assignmentID])

    useEffect(() => {
        if (!state.loadedCourse[courseID.toString()]) {
            actions.getAllCourseSubmissions(courseID)
        }
        return () => {
            actions.setGroupView(false)
            actions.review.setAssignmentID(-1n)
            actions.setActiveEnrollment(null)
        }
    }, [state.loadedCourse])

    if (!state.loadedCourse[courseID.toString()]) {
        return <h1>Fetching Submissions...</h1>
    }

    const generateReviewCell = (submission: Submission, enrollment: Enrollment | Group): RowElement => {
        if (!state.isManuallyGraded(submission)) {
            return { value: "N/A" }
        }
        const reviews = state.review.reviews.get(submission.ID) ?? []
        // Check if the current user has any pending reviews for this submission
        // Used to give cell a box shadow to indicate that the user has a pending review
        const pending = reviews.some((r) => !r.ready && r.ReviewerID === state.self.ID)
        // Check if the this submission is the currently selected submission
        // Used to highlight the cell
        const isSelected = state.activeSubmission === submission.ID
        const score = reviews.reduce((acc, review) => acc + review.score, 0) / reviews.length
        // willBeReleased is true if the average score of all of this submission's reviews is greater than the set minimum score
        // Used to visually indicate that the submission will be released for the given minimum score
        const willBeReleased = state.review.minimumScore > 0 && score >= state.review.minimumScore
        const numReviewers = state.assignments[state.activeCourse.toString()]?.find((a) => a.ID === submission.AssignmentID)?.reviewers ?? 0
        return ({
            // TODO: Figure out a better way to visualize released submissions than '(r)'
            value: `${reviews.length}/${numReviewers} ${submission.released ? "(r)" : ""}`,
            className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""} ${willBeReleased ? "release" : ""} ${pending ? "pending-review" : ""}`,
            onClick: () => {
                actions.setActiveSubmission(submission.ID)
                if (enrollment instanceof Enrollment) {
                    actions.setActiveEnrollment(enrollment.clone())
                }
                actions.setSubmissionOwner({ submission, enrollment })
                actions.review.setSelectedReview(-1)
            }
        })
    }

    const getSubmissionCell = (submission: Submission, enrollment: Enrollment | Group): CellElement => {
        // Check if the this submission is the currently selected submission
        // Used to highlight the cell
        const isSelected = state.activeSubmission === submission.ID
        return ({
            value: `${submission.score} %`,
            className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""}`,
            onClick: () => {
                actions.setActiveSubmission(submission.ID)
                if (enrollment instanceof Enrollment) {
                    actions.setActiveEnrollment(enrollment.clone())
                }
                actions.setSubmissionOwner({ submission, enrollment })
                actions.getSubmission({ submissionID: submission.ID, courseID: state.activeCourse })
            }
        })
    }


    const groupView = state.groupView
    const base: Row = [{ value: "Name", onClick: () => actions.setSubmissionSort(SubmissionSort.Name) }]
    const header = generateAssignmentsHeader(base, assignments, groupView)

    const generator = review ? generateReviewCell : getSubmissionCell
    const rows = generateSubmissionRows(members, generator)


    return (
        <div className="row">
            <div className={state.review.assignmentID >= 0 ? "col-md-4" : "col-xl-6"}>
                <Search placeholder={"Search by name ..."} className="mb-2" >
                    <Button type={ButtonType.BUTTON}
                        classname="ml-2"
                        text={`View by ${groupView ? "student" : "group"}`}
                        onclick={() => { actions.setGroupView(!groupView); actions.review.setAssignmentID(BigInt(-1)) }}
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
