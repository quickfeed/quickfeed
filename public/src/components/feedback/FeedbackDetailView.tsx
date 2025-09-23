import React, { useState } from 'react'
import { useNavigate } from 'react-router'
import { useAppState } from '../../overmind'
import { AssignmentFeedback, Assignment } from '../../../proto/qf/types_pb'
import FeedbackGraph from './FeedbackGraph'
import FeedbackCard from './FeedbackCard'
import FeedbackSortControls from './FeedbackSortControls'

interface FeedbackDetailViewProps {
    assignmentID: string
    assignmentFeedbacks: AssignmentFeedback[]
    assignment?: Assignment
    convertToHoursAndMinutes: (totalMinutes: number) => [number, number]
}

export const FeedbackDetailView: React.FC<FeedbackDetailViewProps> = ({
    assignmentID,
    assignmentFeedbacks,
    assignment,
    convertToHoursAndMinutes
}) => {
    const state = useAppState()
    const navigate = useNavigate()
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc' | 'none'>('none')

    const sortFeedbacks = (feedbacks: AssignmentFeedback[]) => {
        if (sortOrder === 'none') return feedbacks

        return [...feedbacks].sort((a, b) => {
            if (sortOrder === 'asc') {
                return a.TimeSpent - b.TimeSpent
            } else {
                return b.TimeSpent - a.TimeSpent
            }
        })
    }

    const sortedFeedbacks = sortFeedbacks(assignmentFeedbacks)

    return (
        <div className="container mt-4">
            <div className="d-flex align-items-center mb-4">
                <button
                    className="btn btn-outline-secondary mr-3"
                    onClick={() => navigate(`/course/${state.activeCourse}/feedback`)}
                >
                    <i className="fa fa-arrow-left mr-2" /> Back to Summary
                </button>
                <h1 className="text-primary mb-0">
                    <i className="fa fa-comments mr-2" />
                    Feedback for {assignment?.name || `Assignment ${assignmentID}`}
                </h1>
            </div>

            {assignmentFeedbacks.length > 0 ? (
                <>
                    <div className="row">
                        <FeedbackGraph feedbacks={assignmentFeedbacks} />
                    </div>

                    <FeedbackSortControls
                        sortOrder={sortOrder}
                        setSortOrder={setSortOrder}
                        feedbackCount={assignmentFeedbacks.length}
                    />

                    <div className="row">
                        {sortedFeedbacks.map(fb => (
                            <div className="col-lg-4 col-md-6 mb-4" key={fb.ID.toString()}>
                                <FeedbackCard
                                    feedback={fb}
                                    convertToHoursAndMinutes={convertToHoursAndMinutes}
                                />
                            </div>
                        ))}
                    </div>
                </>
            ) : (
                <div className="alert alert-info">
                    <i className="fa fa-info-circle mr-2" />
                    No feedback available for this assignment yet.
                </div>
            )}
        </div>
    )
}

export default FeedbackDetailView
