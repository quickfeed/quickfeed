import { clone, isMessage } from "@bufbuild/protobuf"
import React, { useCallback, useEffect, useRef } from "react"
import { useSearchParams } from 'react-router-dom'
import { Enrollment, EnrollmentSchema, Group, Submission } from "../../proto/qf/types_pb"
import { Color } from "../Helpers"
import { useCourseID } from "../hooks/useCourseID"
import { useActions, useAppState } from "../overmind"
import Button from "./admin/Button"
import TableSort from "./forms/TableSort"
import LabResult from "./LabResult"
import ReviewForm from "./manual-grading/ReviewForm"
import Search from "./Search"
import { SubmissionsTable } from "./submissions-table"

const Results = ({ review }: { review: boolean }) => {
    const state = useAppState()
    const actions = useActions()
    const courseID = useCourseID()
    const [searchParams, setSearchParams] = useSearchParams()

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
            return
        }
        const selectedLab = searchParams.get("id")
        if (selectedLab) {
            const submission = state.submissionsForCourse.ByID(BigInt(selectedLab))
            if (submission) {
                actions.global.setSelectedSubmission({ submission })
                actions.global.updateSubmissionOwner(state.submissionsForCourse.OwnerByID(submission.ID))
                if (submission.reviews.length > 0) {
                    actions.review.setSelectedReview(-1)
                } else {
                    actions.global.getSubmission({ submission, owner: state.submissionOwner, courseID: state.activeCourse })
                }
            }
        }
    }, [loaded])

    const handleSubmissionClick = useCallback((submission: Submission, owner: Enrollment | Group) => {
        actions.global.setSelectedSubmission({ submission })
        if (isMessage(owner, EnrollmentSchema)) {
            actions.global.setActiveEnrollment(clone(EnrollmentSchema, owner))
        }
        actions.global.setSubmissionOwner(owner)
        setSearchParams({ id: submission.ID.toString() })

        if (review) {
            actions.review.setSelectedReview(-1)
        } else {
            actions.global.getSubmission({ submission, owner: state.submissionOwner, courseID: state.activeCourse })
        }
    }, [actions, review, setSearchParams, state.activeCourse, state.submissionOwner])

    const toggleGroupView = useCallback(() => {
        actions.global.setGroupView(!state.groupView)
        actions.review.setAssignmentID(-1n)
    }, [actions, state.groupView])

    if (!loaded) {
        return (
            <div className="flex justify-center items-center py-12">
                <span className="loading loading-spinner loading-lg" />
            </div>
        )
    }

    const gridCols = state.review.assignmentID >= 0 ? "lg:grid-cols-3" : "lg:grid-cols-2"
    const displayMode = state.groupView ? "Group" : "Student"
    const buttonColor = state.groupView ? Color.BLUE : Color.GREEN

    return (
        <div className={`grid grid-cols-1 ${gridCols} gap-6`}>
            <div className="space-y-4">
                <Search placeholder="Search by name..." className="mb-2">
                    <Button
                        text={`View by ${displayMode}`}
                        color={buttonColor}
                        className="ml-2"
                        onClick={toggleGroupView}
                    />
                </Search>
                <TableSort />
                <div className="overflow-x-auto">
                    <SubmissionsTable onSubmissionClick={handleSubmissionClick} review={review} />
                </div>
            </div >
            <div className={state.review.assignmentID >= 0 ? "lg:col-span-2" : ""}>
                {review ? <ReviewForm /> : <LabResult />}
            </div>
        </div >
    )
}

export default Results
