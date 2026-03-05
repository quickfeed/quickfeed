import React from 'react'
import { useNavigate } from 'react-router'
import { useAppState } from '../../overmind'
import { Assignment } from '../../../proto/qf/types_pb'
import { CourseFeedbackData } from "../../overmind/namespaces/feedback/state"

interface FeedbackSummaryViewProps {
    assignments: Assignment[]
    courseFeedbackData: CourseFeedbackData | undefined
    avgTimeSpent: (assignmentID: bigint) => [number, number]
}

export const FeedbackSummaryView: React.FC<FeedbackSummaryViewProps> = ({
    assignments,
    courseFeedbackData,
    avgTimeSpent
}) => {
    const state = useAppState()
    const navigate = useNavigate()

    return (
        <div className="container mx-auto px-4 py-8 max-w-7xl">
            <div className="mb-8">
                <div className="flex items-center gap-4 mb-2">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary to-secondary flex items-center justify-center shadow-lg">
                        <i className="fa fa-comments text-white text-xl" />
                    </div>
                    <h1 className="text-4xl font-bold text-base-content">Assignment Feedback Summary</h1>
                </div>
                <p className="text-base-content/60 ml-16">Review student feedback for all assignments</p>
            </div>

            {assignments && courseFeedbackData ? (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {assignments.map(assignment => {
                        const feedbacks = courseFeedbackData.byAssignment.get(assignment.ID) || []
                        const [avgHours, avgMinutes] = avgTimeSpent(assignment.ID)

                        return (
                            <div
                                key={assignment.ID.toString()}
                                className="card bg-base-100 shadow-lg hover:shadow-2xl transition-all duration-300 cursor-pointer hover:-translate-y-1"
                                onClick={() => navigate(`/course/${state.activeCourse}/feedback/${assignment.ID}`)}
                                role="button"
                                aria-hidden="true"
                            >
                                <div className="card-body p-6">
                                    <div className="flex items-start gap-3 mb-4">
                                        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0">
                                            <i className="fa fa-book text-primary" />
                                        </div>
                                        <div className="flex-1">
                                            <h5 className="text-lg font-bold text-base-content mb-1">{assignment.name}</h5>
                                            <p className="text-sm text-base-content/60">Assignment Feedback</p>
                                        </div>
                                    </div>

                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="bg-info/10 rounded-lg p-4 text-center border border-info/20">
                                            <div className="text-3xl font-bold text-info mb-1">{feedbacks.length}</div>
                                            <div className="text-xs text-base-content/70 uppercase tracking-wide">Responses</div>
                                        </div>
                                        <div className="bg-success/10 rounded-lg p-4 text-center border border-success/20">
                                            <div className="text-3xl font-bold text-success mb-1">
                                                {(avgHours > 0 || avgMinutes > 0) ? `${avgHours}h ${avgMinutes}m` : 'N/A'}
                                            </div>
                                            <div className="text-xs text-base-content/70 uppercase tracking-wide">Avg. Time</div>
                                        </div>
                                    </div>

                                    {feedbacks.length > 0 && (
                                        <div className="mt-4 pt-4 border-t border-base-300">
                                            <div className="flex items-center justify-center gap-2 text-sm text-base-content/60">
                                                <i className="fa fa-hand-pointer-o" />
                                                <span>Click to view detailed feedback</span>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </div>
                        )
                    })}
                </div>
            ) : (
                <div className="alert alert-info shadow-lg">
                    <div className="flex items-center gap-2">
                        <i className="fa fa-info-circle text-xl" />
                        <span>No feedback available for this course yet.</span>
                    </div>
                </div>
            )}
        </div>
    )
}

export default FeedbackSummaryView
