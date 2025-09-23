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
        <div className="container mt-4">
            <h1 className="mb-4 text-primary">
                <i className="fa fa-comments mr-2"></i>Assignment Feedback Summary
            </h1>
            {assignments && courseFeedbackData ? (
                <div className="row">
                    {assignments.map(assignment => {
                        const feedbacks = courseFeedbackData.byAssignment.get(assignment.ID) || []
                        const [avgHours, avgMinutes] = avgTimeSpent(assignment.ID)

                        return (
                            <div className="col-lg-6 col-md-6 mb-4" key={assignment.ID.toString()}>
                                <div
                                    className="card shadow-sm h-100 cursor-pointer"
                                    onClick={() => navigate(`/course/${state.activeCourse}/feedback/${assignment.ID}`)}
                                    style={{ cursor: 'pointer' }}
                                >
                                    <div className="card-header bg-primary text-white">
                                        <h5 className="mb-0 d-flex justify-content-between align-items-center">
                                            <span>
                                                <i className="fa fa-book mr-2"></i>
                                                {assignment.name}
                                            </span>
                                        </h5>
                                    </div>
                                    <div className="card-body">
                                        <div className="row text-center">
                                            <div className="col-6">
                                                <div className="h3 text-info mb-0">{feedbacks.length}</div>
                                                <small className="text-muted">Responses</small>
                                            </div>
                                            <div className="col-6">
                                                <div className="h3 text-success mb-0">
                                                    {(avgHours > 0 || avgMinutes > 0) ? `${avgHours}h ${avgMinutes}m` : 'N/A'}
                                                </div>
                                                <small className="text-muted">Avg. Time</small>
                                            </div>
                                        </div>
                                        {feedbacks.length > 0 && (
                                            <div className="mt-3">
                                                <small className="text-muted">
                                                    <i className="fa fa-hand-pointer-o mr-1"></i>
                                                    Click to view detailed feedback
                                                </small>
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )
                    })}
                </div>
            ) : (
                <div className="alert alert-info">
                    <i className="fa fa-info-circle mr-2"></i>
                    No feedback available for this course yet.
                </div>
            )}
        </div>
    )
}

export default FeedbackSummaryView
