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
        <div className="container mx-auto px-4 py-8 max-w-7xl">
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 mb-8">
                <button
                    className="btn btn-ghost gap-2 hover:btn-primary"
                    onClick={() => navigate(`/course/${state.activeCourse}/feedback`)}
                >
                    <i className="fa fa-arrow-left" />
                    Back to Summary
                </button>
                <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary to-secondary flex items-center justify-center shadow-lg">
                        <i className="fa fa-comments text-white text-xl" />
                    </div>
                    <div>
                        <h1 className="text-3xl font-bold text-base-content">
                            {assignment?.name || `Assignment ${assignmentID}`}
                        </h1>
                        <p className="text-sm text-base-content/60">Detailed Feedback View</p>
                    </div>
                </div>
            </div>

            {assignmentFeedbacks.length > 0 ? (
                <>
                    <div className="mb-6">
                        <FeedbackGraph feedbacks={assignmentFeedbacks} />
                    </div>

                    <FeedbackSortControls
                        sortOrder={sortOrder}
                        setSortOrder={setSortOrder}
                        feedbackCount={assignmentFeedbacks.length}
                    />

                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {sortedFeedbacks.map(fb => (
                            <FeedbackCard
                                key={fb.ID.toString()}
                                feedback={fb}
                                convertToHoursAndMinutes={convertToHoursAndMinutes}
                            />
                        ))}
                    </div>
                </>
            ) : (
                <div className="alert alert-info shadow-lg">
                    <div className="flex items-center gap-2">
                        <i className="fa fa-info-circle text-xl" />
                        <span>No feedback available for this assignment yet.</span>
                    </div>
                </div>
            )}
        </div>
    )
}

export default FeedbackDetailView
