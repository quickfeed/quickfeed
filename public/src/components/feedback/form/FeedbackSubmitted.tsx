import React from 'react'
import { Assignment } from '../../../../proto/qf/types_pb'

interface FeedbackSubmittedCardProps {
    assignment: Assignment
}

export const FeedbackSubmittedCard: React.FC<FeedbackSubmittedCardProps> = ({ assignment }) => {
    return (
        <div className="card mt-3">
            <div className="card-body">
                <h5 className="card-title text-success">
                    <i className="fa fa-check-circle me-2" />
                    Feedback Submitted
                </h5>
                <p className="card-text">Thank you for your feedback on {assignment.name}!</p>
            </div>
        </div>
    )
}

export default FeedbackSubmittedCard
