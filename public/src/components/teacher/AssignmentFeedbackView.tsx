import React, { useEffect } from "react"
import { useParams, useNavigate } from "react-router"
import { useActions, useAppState } from "../../overmind"

export const AssignmentFeedbackView = () => {
    const state = useAppState()
    const actions = useActions()
    const navigate = useNavigate()
    const { assignmentID } = useParams<{ assignmentID: string }>()

    useEffect(() => {
        const fetchFeedback = async () => {
            await actions.feedback.getAssignmentFeedback({ courseID: state.activeCourse, category: "assignmentID", categoryValue: BigInt(0) })
        }

        fetchFeedback()
    }, [actions.feedback, state.activeCourse])

    const courseFeedbackData = state.feedback.feedback.get(state.activeCourse)
    const assignments = state.assignments[state.activeCourse.toString()]

    // If assignmentID is provided, show detailed view for that assignment
    if (assignmentID) {
        const assignmentIDBigInt = BigInt(assignmentID)
        const assignmentFeedbacks = courseFeedbackData?.byAssignment.get(assignmentIDBigInt) || []
        const assignment = assignments?.find(a => a.ID === assignmentIDBigInt)

        return (
            <div className="container mt-4">
                <div className="d-flex align-items-center mb-4">
                    <button
                        className="btn btn-outline-secondary mr-3"
                        onClick={() => navigate(`/course/${state.activeCourse}/feedback`)}
                    >
                        <i className="fa fa-arrow-left mr-2"></i>Back to Summary
                    </button>
                    <h1 className="text-primary mb-0">
                        <i className="fa fa-comments mr-2"></i>
                        Feedback for {assignment?.name || `Assignment ${assignmentID}`}
                    </h1>
                </div>

                {assignmentFeedbacks.length > 0 ? (
                    <div className="row">
                        {assignmentFeedbacks.map(fb => (
                            <div className="col-md-6 mb-4" key={fb.ID}>
                                <div className="card shadow-sm h-100">
                                    <div className="card-header bg-info text-white">
                                        <h5 className="mb-0">
                                            <i className="fa fa-user mr-2"></i>
                                            Feedback #{fb.ID.toString()}
                                            {fb.TimeSpent > 0 && (
                                                <span className="badge badge-pill badge-light ml-2">{fb.TimeSpent}h</span>
                                            )}
                                        </h5>
                                    </div>
                                    <div className="card-body">
                                        <div className="mb-3">
                                            <h6 className="text-info">
                                                <i className="fa fa-heart mr-2"></i>What they liked:
                                            </h6>
                                            <p className="card-text">{fb.LikedContent || <span className="text-muted">No feedback provided</span>}</p>
                                        </div>
                                        <div className="mb-3">
                                            <h6 className="text-info">
                                                <i className="fa fa-lightbulb mr-2"></i>Suggestions for improvement:
                                            </h6>
                                            <p className="card-text">{fb.ImprovementSuggestions || <span className="text-muted">No suggestions provided</span>}</p>
                                        </div>
                                        <div className="d-flex justify-content-between">
                                            {fb.UserID === 0n ? (
                                                <span className="badge badge-warning">Anonymous</span>
                                            ) : (
                                                <span className="badge badge-info">User ID: {fb.UserID.toString()}</span>
                                            )}
                                            {fb.CreatedAt && (
                                                <small className="text-muted">
                                                    {new Date(Number(fb.CreatedAt.seconds) * 1000).toLocaleDateString()}
                                                </small>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="alert alert-info">
                        <i className="fa fa-info-circle mr-2"></i>
                        No feedback available for this assignment yet.
                    </div>
                )}
            </div>
        )
    }

    // Summary view - show all assignments with feedback counts
    return (
        <div className="container mt-4">
            <h1 className="mb-4 text-primary">
                <i className="fa fa-comments mr-2"></i>Assignment Feedback Summary
            </h1>
            {assignments && courseFeedbackData ? (
                <div className="row">
                    {assignments.map(assignment => {
                        const feedbacks = courseFeedbackData.byAssignment.get(assignment.ID) || []
                        const avgTimeSpent = feedbacks.length > 0
                            ? feedbacks.reduce((sum, fb) => sum + fb.TimeSpent, 0) / feedbacks.length
                            : 0

                        return (
                            <div className="col-md-6 mb-4" key={assignment.ID.toString()}>
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
                                                    {avgTimeSpent > 0 ? `${avgTimeSpent.toFixed(1)}h` : 'N/A'}
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

export default AssignmentFeedbackView
