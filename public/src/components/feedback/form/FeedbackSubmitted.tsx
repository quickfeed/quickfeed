import React from 'react'
import { Assignment } from '../../../../proto/qf/types_pb'

interface FeedbackSubmittedCardProps {
    assignment: Assignment
}

export const FeedbackSubmittedCard: React.FC<FeedbackSubmittedCardProps> = ({ assignment }) => {
    return (
        <div className="alert alert-success shadow-lg my-6">
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
                <div className="w-12 h-12 rounded-full bg-success/20 flex items-center justify-center flex-shrink-0">
                    <i className="fa fa-check-circle text-2xl" />
                </div>
                <div>
                    <h5 className="font-bold text-lg mb-1">Feedback Submitted Successfully!</h5>
                    <p className="text-sm">Thank you for your feedback on {assignment.name}!</p>
                </div>
            </div>
        </div>
    )
}

export default FeedbackSubmittedCard
