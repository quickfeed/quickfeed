import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { useActions, useAppState } from '../../overmind'
import FeedbackSummaryView from './FeedbackSummaryView'
import FeedbackDetailView from './FeedbackDetailView'
import { convertToBigInt } from "../../Helpers"

export const AssignmentFeedbackView = () => {
    const state = useAppState()
    const actions = useActions()
    const { assignmentID } = useParams<{ assignmentID: string }>()

    useEffect(() => {
        const fetchFeedback = async () => {
            await actions.feedback.getAssignmentFeedback({
                courseID: state.activeCourse,
                category: "assignmentID",
                categoryValue: BigInt(0)
            })
        }

        void fetchFeedback()
    }, [actions.feedback, state.activeCourse])

    const courseFeedbackData = state.feedback.feedback.get(state.activeCourse)
    const assignments = state.assignments[state.activeCourse.toString()]

    const avgTimeSpent = (assignmentID: bigint): [number, number] => {
        const feedbacks = courseFeedbackData?.byAssignment.get(assignmentID) || []
        const numWithTimeSpent = feedbacks.filter(fb => fb.TimeSpent > 0).length
        const total = feedbacks.reduce((acc, fb) => {
            return acc + (fb.TimeSpent || 0)
        }, 0)

        return convertToHoursAndMinutes(total / numWithTimeSpent)
    }

    const convertToHoursAndMinutes = (totalMinutes: number): [number, number] => {
        const hours = Math.floor(totalMinutes / 60)
        const minutes = Math.floor(totalMinutes % 60)
        return [hours, minutes]
    }

    // If assignmentID is provided, show detailed view for that assignment
    if (assignmentID) {
        const assignmentIDBigInt = convertToBigInt(assignmentID)
        const assignmentFeedbacks = courseFeedbackData?.byAssignment.get(assignmentIDBigInt) || []
        const assignment = assignments?.find(a => a.ID === assignmentIDBigInt)

        return (
            <FeedbackDetailView
                assignmentID={assignmentID}
                assignmentFeedbacks={assignmentFeedbacks}
                assignment={assignment}
                convertToHoursAndMinutes={convertToHoursAndMinutes}
            />
        )
    }

    // Summary view - show all assignments with feedback counts
    return (
        <FeedbackSummaryView
            assignments={assignments}
            courseFeedbackData={courseFeedbackData}
            avgTimeSpent={avgTimeSpent}
        />
    )
}

export default AssignmentFeedbackView
