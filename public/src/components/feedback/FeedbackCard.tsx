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
        <div className="card bg-base-200 shadow-lg hover:shadow-xl transition-shadow h-full">
            <div className="card-body">
                <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                        <div className="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
                            <i className="fa fa-user text-primary" />
                        </div>
                        <h5 className="font-semibold text-base-content">
                            Feedback #{feedback.ID.toString()}
                        </h5>
                    </div>
                    {feedback.TimeSpent > 0 && (
                        <div className="badge badge-primary badge-lg gap-2">
                            <i className="fa fa-clock-o text-xs" />
                            {(() => {
                                const [hours, minutes] = convertToHoursAndMinutes(feedback.TimeSpent)
                                if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`
                                if (hours > 0) return `${hours}h`
                                return `${minutes}m`
                            })()}
                        </div>
                    )}
                </div>

                <div className="space-y-4">
                    <div className="bg-success/5 rounded-lg p-4 border-l-4 border-success">
                        <div className="flex items-center gap-2 mb-2">
                            <i className="fa fa-heart text-success" />
                            <h6 className="font-semibold text-success">What they liked</h6>
                        </div>
                        <p className="text-sm text-base-content/80 leading-relaxed">
                            {feedback.LikedContent || <span className="text-base-content/50 italic">No feedback provided</span>}
                        </p>
                    </div>

                    <div className="bg-warning/5 rounded-lg p-4 border-l-4 border-warning">
                        <div className="flex items-center gap-2 mb-2">
                            <i className="fa fa-lightbulb-o text-warning" />
                            <h6 className="font-semibold text-warning">Suggestions for improvement</h6>
                        </div>
                        <p className="text-sm text-base-content/80 leading-relaxed">
                            {feedback.ImprovementSuggestions || <span className="text-base-content/50 italic">No suggestions provided</span>}
                        </p>
                    </div>
                </div>

                {feedback.CreatedAt && (
                    <div className="flex justify-end mt-4 pt-4 border-t border-base-300">
                        <span className="text-xs text-base-content/60 flex items-center gap-1">
                            <i className="fa fa-calendar" />
                            {new Date(Number(feedback.CreatedAt.seconds) * 1000).toLocaleDateString()}
                        </span>
                    </div>
                )}
            </div>
        </div>
    )
}

export default FeedbackCard
