import React, { useCallback, useEffect, useMemo } from "react"
import { useHistory, useLocation } from 'react-router-dom';
import { Enrollment, Group, Submission } from "../../proto/qf/types_pb"
import { Color, getCourseID, getSubmissionCellColor } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Button, { ButtonType } from "./admin/Button"
import { generateAssignmentsHeader, generateSubmissionRows } from "./ComponentsHelpers"
import DynamicTable, { CellElement, RowElement } from "./DynamicTable"
import TableSort from "./forms/TableSort"
import LabResult from "./LabResult"
import ReviewForm from "./manual-grading/ReviewForm"
import Release from "./Release"
import Search from "./Search"


const Results = ({ review }: { review: boolean }): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const history = useHistory();
    const location = useLocation();

    const members = useMemo(() => { return state.courseMembers }, [state.courseMembers, state.groupView])
    const assignments = useMemo(() => {
        // Filter out all assignments that are not the selected assignment, if any assignment is selected
        return state.assignments[courseID.toString()]?.filter(a => state.review.assignmentID <= 0 || a.ID === state.review.assignmentID) ?? []
    }, [state.assignments, courseID, state.review.assignmentID])

    useEffect(() => {
        if (!state.loadedCourse[courseID.toString()]) {
            actions.loadCourseSubmissions(courseID)
        }
        return () => {
            actions.setGroupView(false)
            actions.review.setAssignmentID(-1n)
            actions.setActiveEnrollment(null)
        }
    }, [])

    useEffect(() => {
        if (!state.selectedSubmission) {
            // If no submission is selected, check if there is a selected lab in the URL
            // and select it if it exists
            const selectedLab = new URLSearchParams(location.search).get('id');
            if (selectedLab) {
                const submission = state.submissionsForCourse.ByID(BigInt(selectedLab))
                if (submission) {
                    actions.setSelectedSubmission(submission)
                    actions.updateSubmissionOwner(state.submissionsForCourse.OwnerByID(submission.ID))
                }
            }
        }
    }, [])

    const handleLabClick = useCallback((labId: bigint) => {
        // Update the URL with the selected lab
        history.replace({
            pathname: location.pathname,
            search: `?id=${labId}`
        });
    }, [history]);

    if (!state.loadedCourse[courseID.toString()]) {
        return <h1>Fetching Submissions...</h1>
    }

    const generateReviewCell = (submission: Submission, owner: Enrollment | Group): RowElement => {
        if (!state.isManuallyGraded(submission)) {
            return { value: "N/A" }
        }
        const reviews = state.review.reviews.get(submission.ID) ?? []
        // Check if the current user has any pending reviews for this submission
        // Used to give cell a box shadow to indicate that the user has a pending review
        const pending = reviews.some((r) => !r.ready && r.ReviewerID === state.self.ID)
        // Check if the this submission is the currently selected submission
        // Used to highlight the cell
        const isSelected = state.selectedSubmission?.ID === submission.ID
        const score = reviews.reduce((acc, theReview) => acc + theReview.score, 0) / reviews.length
        // willBeReleased is true if the average score of all of this submission's reviews is greater than the set minimum score
        // Used to visually indicate that the submission will be released for the given minimum score
        const willBeReleased = state.review.minimumScore > 0 && score >= state.review.minimumScore
        const numReviewers = state.assignments[state.activeCourse.toString()]?.find((a) => a.ID === submission.AssignmentID)?.reviewers ?? 0
        return ({
            // TODO: Figure out a better way to visualize released submissions than '(r)'
            value: `${reviews.length}/${numReviewers} ${submission.released ? "(r)" : ""}`,
            className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""} ${willBeReleased ? "release" : ""} ${pending ? "pending-review" : ""}`,
            onClick: () => {
                actions.setSelectedSubmission(submission)
                if (owner instanceof Enrollment) {
                    actions.setActiveEnrollment(owner.clone())
                }
                actions.setSubmissionOwner(owner)
                actions.review.setSelectedReview(-1)
                handleLabClick(submission.ID)
            }
        })
    }

    const getSubmissionCell = (submission: Submission, owner: Enrollment | Group): CellElement => {
        // Check if the this submission is the currently selected submission
        // Used to highlight the cell
        const isSelected = state.selectedSubmission?.ID === submission.ID
        return ({
            value: `${submission.score} %`,
            className: `${getSubmissionCellColor(submission)} ${isSelected ? "selected" : ""}`,
            onClick: () => {
                actions.setSelectedSubmission(submission)
                if (owner instanceof Enrollment) {
                    actions.setActiveEnrollment(owner.clone())
                }
                actions.setSubmissionOwner(owner)
                handleLabClick(submission.ID)
                actions.getSubmission({ submission: submission, owner: state.submissionOwner, courseID: state.activeCourse })
            }
        })
    }

    const groupView = state.groupView
    const header = generateAssignmentsHeader(assignments, groupView)

    const generator = review ? generateReviewCell : getSubmissionCell
    const rows = generateSubmissionRows(members, generator)


    return (
        <div className="row">
            <div className={`p-0 ${state.review.assignmentID >= 0 ? "col-md-4" : "col-md-6"}`}>
                {review ? <Release /> : null}
                <Search placeholder={"Search by name ..."} className="mb-2" >
                    <Button
                        text={`View by ${groupView ? "student" : "group"}`}
                        color={groupView ? Color.BLUE : Color.GREEN}
                        type={ButtonType.BUTTON}
                        className="ml-2"
                        onClick={() => { actions.setGroupView(!groupView); actions.review.setAssignmentID(BigInt(-1)) }}
                    />
                </Search>
                <TableSort review={review} />
                <DynamicTable header={header} data={rows} />
            </div>
            <div className="col">
                {review ? <ReviewForm /> : <LabResult />}
            </div>
        </div>
    )
}

export default Results
