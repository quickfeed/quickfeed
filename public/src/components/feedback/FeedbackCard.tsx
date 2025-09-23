import React from 'react'
import { AssignmentFeedback } from '../../../proto/qf/types_pb'

interface FeedbackCardProps {
    feedback: AssignmentFeedback
    convertToHoursAndMinutes: (totalMinutes: number) => [number, number]
}

export const FeedbackCard: React.FC<FeedbackCardProps> = ({
    feedback,
    convertToHoursAndMinutes
}) => {
    return (
        <div className="card shadow-sm h-100">
            <div className="card-header bg-primary text-white">
                <h5 className="mb-0">
                    <i className="fa fa-user mr-2"></i>
                    Feedback #{feedback.ID.toString()}
                    {feedback.TimeSpent > 0 && (
                        <span className="badge badge-pill badge-light ml-2">
                            {(() => {
                                const [hours, minutes] = convertToHoursAndMinutes(feedback.TimeSpent)
                                if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`
                                if (hours > 0) return `${hours}h`
                                return `${minutes}m`
                            })()}
                        </span>
                    )}
                </h5>
            </div>
            <div className="card-body">
                <div className="mb-3">
                    <h6 className="text-primary">
                        <i className="fa fa-heart mr-2"></i>What they liked:
                    </h6>
                    <p className="card-text">
                        {feedback.LikedContent || <span className="text-muted">No feedback provided</span>}
                    </p>
                </div>
                <div className="mb-3">
                    <h6 className="text-primary">
                        <i className="fa fa-lightbulb mr-2"></i>Suggestions for improvement:
                    </h6>
                    <p className="card-text">
                        {feedback.ImprovementSuggestions || <span className="text-muted">No suggestions provided</span>}
                    </p>
                </div>
                <div className="d-flex justify-content-between">
                    {feedback.UserID === 0n ? (
                        <span className="badge badge-warning">Anonymous</span>
                    ) : (
                        <span className="badge badge-info">User ID: {feedback.UserID.toString()}</span>
                    )}
                    {feedback.CreatedAt && (
                        <small className="text-muted">
                            {new Date(Number(feedback.CreatedAt.seconds) * 1000).toLocaleDateString()}
                        </small>
                    )}
                </div>
            </div>
        </div>
    )
}

export default FeedbackCard
