import { clone, isMessage } from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useMemo, useRef } from "react"
import { useSearchParams } from 'react-router-dom'
import { Enrollment, EnrollmentSchema, Group, Submission } from "../../proto/qf/types_pb"
import { Color, getSubmissionCellColor, Icon } from "../Helpers"
import { useCourseID } from "../hooks/useCourseID"
import { useActions, useAppState } from "../overmind"
import Button, { ButtonType } from "./admin/Button"
import { generateAssignmentsHeader, generateSubmissionRows } from "./ComponentsHelpers"
import DynamicTable, { CellElement, RowElement } from "./DynamicTable"
import TableSort from "./forms/TableSort"
import LabResult from "./LabResult"
import ReviewForm from "./manual-grading/ReviewForm"
import Search from "./Search"

const Results = ({ review }: { review: boolean }) => {
    const state = useAppState()
    const actions = useActions()
    const courseID = useCourseID()
    const [searchParams, setSearchParams] = useSearchParams()

    const members = useMemo(() => { return state.courseMembers }, [state.courseMembers])
    const assignments = useMemo(() => {
        // Filter out all assignments that are not the selected assignment, if any assignment is selected
        return state.assignments[courseID.toString()]?.filter(
            a => state.review.assignmentID <= 0 || a.ID === state.review.assignmentID
        )
    }, [state.assignments, courseID, state.review.assignmentID])
    const loaded = state.loadedCourse[courseID.toString()]

    // Always keep latest state/actions/searchParams in a ref for effects
    const latest = useRef({ state, actions, searchParams })
    useEffect(() => {
        latest.current = { state, actions, searchParams }
    }, [state, actions, searchParams])

    // Load the course submissions when the component mounts
    useEffect(() => {
        if (!state.loadedCourse[courseID.toString()]) {
            actions.global.loadCourseSubmissions(courseID)
        }
        return () => {
            actions.global.setGroupView(false)
            actions.review.setAssignmentID(-1n)
            actions.global.setActiveEnrollment(null)
            actions.global.setSelectedSubmission({ submission: null })
        }
    }, [actions, courseID, state.loadedCourse])

    // Select submission from URL if not already selected, after loading is done
    useEffect(() => {
        const { state, actions, searchParams } = latest.current
        if (state.selectedSubmission) {
            // submission is already selected, nothing to do
            return
        }
        // If no submission is selected, check if there is a selected lab in the URL
        // and select it if it exists
        const selectedLab = searchParams.get("id")
        if (selectedLab) {
            const submission = state.submissionsForCourse.ByID(BigInt(selectedLab))
            if (submission) {
                actions.global.setSelectedSubmission({ submission })
                actions.global.updateSubmissionOwner(state.submissionsForCourse.OwnerByID(submission.ID))
                if (submission.reviews.length > 0) {
                    // If the submission has reviews we need to set the selected review to -1
                    // to show the review form
                    actions.review.setSelectedReview(-1)
                } else {
                    // Fetch full submission data since the submission data by default does not include the build log
                    actions.global.getSubmission({ submission, owner: state.submissionOwner, courseID: state.activeCourse })
                }
            }
        }
    }, [loaded])

    const groupView = state.groupView
    const handleSetGroupView = useCallback(() => {
        actions.global.setGroupView(!groupView)
        actions.review.setAssignmentID(BigInt(-1))
    }, [actions, groupView])

    const handleLabClick = useCallback((submission: Submission, owner: Enrollment | Group) => {
        actions.global.setSelectedSubmission({ submission })
        if (isMessage(owner, EnrollmentSchema)) {
            actions.global.setActiveEnrollment(clone(EnrollmentSchema, owner))
        }
        actions.global.setSubmissionOwner(owner)
        // Update the URL with the selected lab
        setSearchParams({ id: submission.ID.toString() })
    }, [actions, setSearchParams])

    const handleReviewCellClick = useCallback((submission: Submission, owner: Enrollment | Group) => () => {
        handleLabClick(submission, owner)
        actions.review.setSelectedReview(-1)
    }, [actions, handleLabClick])

    const handleSubmissionCellClick = useCallback((submission: Submission, owner: Enrollment | Group) => () => {
        handleLabClick(submission, owner)
        actions.global.getSubmission({ submission, owner: state.submissionOwner, courseID: state.activeCourse })
    }, [actions, handleLabClick, state.activeCourse, state.submissionOwner])

    if (!state.loadedCourse[courseID.toString()]) {
        return <h1>Fetching Submissions...</h1>
    }

    const generateReviewCell = (submission: Submission, owner: Enrollment | Group): RowElement => {
        if (!state.isManuallyGraded(submission)) {
            return { iconTitle: "auto graded", iconClassName: Icon.DASH, value: "" }
        }
        const reviews = state.review.reviews.get(submission.ID) ?? []
        // Check if the this submission is the currently selected submission
        // Used to highlight the cell
        const isSelected = state.selectedSubmission?.ID === submission.ID ? "selected" : ""
        const numReviewers = state.assignments[state.activeCourse.toString()]?.find((a) => a.ID === submission.AssignmentID)?.reviewers ?? 0
        return ({
            value: `${reviews.length}/${numReviewers}`,
            className: `${getSubmissionCellColor(submission, owner)} ${isSelected}`,
            onClick: handleReviewCellClick(submission, owner),
        })
    }

    const getSubmissionCell = (submission: Submission, owner: Enrollment | Group): CellElement => {
        // Check if the this submission is the currently selected submission
        // Used to highlight the cell
        const isSelected = state.selectedSubmission?.ID === submission.ID ? "selected" : ""
        return ({
            value: `${submission.score} %`,
            className: `${getSubmissionCellColor(submission, owner)} ${isSelected}`,
            onClick: handleSubmissionCellClick(submission, owner),
        })
    }

    const header = generateAssignmentsHeader(assignments, groupView, actions, state.isCourseManuallyGraded)

    const generator = review ? generateReviewCell : getSubmissionCell
    const rows = generateSubmissionRows(members, generator, state)

    const displayMode = state.groupView ? "Group" : "Student"
    const buttonColor = state.groupView ? Color.BLUE : Color.GREEN
    return (
        <div className="row">
            <div className="p-0 col-md-6">
                <Search placeholder={"Search by name ..."} className="mb-2" >
                    <Button
                        text={`View by ${displayMode}`}
                        color={buttonColor}
                        type={ButtonType.BUTTON}
                        className="ml-2"
                        onClick={handleSetGroupView}
                    />
                </Search>
                <TableSort />
                <DynamicTable header={header} data={rows} />
            </div>
            <div className="col-md-6">
                {review ? <ReviewForm /> : <LabResult />}
            </div>
        </div>
    )
}

export default Results
