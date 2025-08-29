import { create } from "@bufbuild/protobuf"
import React, { useState } from 'react'
import { Assignment, AssignmentFeedback, AssignmentFeedbackSchema } from '../../../proto/qf/types_pb'
import { Color } from "../../Helpers"
import { useActions, useAppState } from '../../overmind'

interface AssignmentFeedbackFormProps {
    assignment: Assignment
    courseID: string
}

const AssignmentFeedbackForm: React.FC<AssignmentFeedbackFormProps> = ({ assignment, courseID }) => {
    const state = useAppState()
    const actions = useActions()
    const [isOpen, setIsOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [isSubmitted, setIsSubmitted] = useState(false)
    const [likedContent, setLikedContent] = useState('')
    const [improvementSuggestions, setImprovementSuggestions] = useState('')
    const [timeSpent, setTimeSpent] = useState('')
    const [anonymous, setAnonymous] = useState(true)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (likedContent.trim().length < 10 && improvementSuggestions.trim().length < 10) {
            actions.global.alert({ text: 'Please provide at least 10 words of feedback', color: Color.YELLOW })
            return
        }

        if (likedContent.length > 200 || improvementSuggestions.length > 200 || timeSpent.length > 30) {
            actions.global.alert({ text: 'Please keep responses under the 200 word limit', color: Color.YELLOW })
            return
        }

        setIsSubmitting(true)

        try {
            const feedback: AssignmentFeedback = create(AssignmentFeedbackSchema, {
                ID: BigInt(0), // Will be set by backend
                CourseID: assignment.CourseID,
                AssignmentID: assignment.ID,
                UserID: anonymous ? BigInt(0) : state.self.ID, // Will be set by backend if not anonymous
                LikedContent: likedContent.trim(),
                ImprovementSuggestions: improvementSuggestions.trim(),
                TimeSpent: timeSpent.trim(),
            })

            await actions.feedback.createAssignmentFeedback({ courseID, feedback })
            setIsSubmitted(true)
            setIsOpen(false)

            // Reset form
            setLikedContent('')
            setImprovementSuggestions('')
            setTimeSpent('')
        } catch (error) {
            console.error('Failed to submit feedback:', error)
            actions.global.alert({ text: 'Failed to submit feedback. Please try again.', color: Color.RED })
        } finally {
            setIsSubmitting(false)
        }
    }

    if (isSubmitted) {
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

    return (
        <div className="card mt-3">
            <div className="card-header">
                <button
                    className="btn btn-link p-0 text-decoration-none w-100 text-start"
                    onClick={() => setIsOpen(!isOpen)}
                    type="button"
                    aria-expanded={isOpen}
                >
                    <h5 className="mb-0">
                        <i className={`fa fa-chevron-${isOpen ? 'down' : 'right'} me-2`} />
                        Give Feedback on This Assignment
                    </h5>
                </button>
            </div>

            {isOpen && (
                <div className="card-body">
                    <form onSubmit={handleSubmit}>
                        <div className="mb-3">
                            <label htmlFor="likedContent" className="form-label">
                                What did you like about this assignment? <small className="text-muted">(min 10 words, max 200 words)</small>
                            </label>
                            <textarea
                                id="likedContent"
                                className="form-control"
                                rows={3}
                                value={likedContent}
                                onChange={(e) => setLikedContent(e.target.value)}
                                placeholder="What worked well? What was interesting or helpful?"
                                maxLength={200}
                            />
                            <small className="form-text text-muted">
                                {likedContent.length}/200 characters
                            </small>
                        </div>

                        <div className="mb-3">
                            <label htmlFor="improvementSuggestions" className="form-label">
                                What would make this assignment better? <small className="text-muted">(min 10 words, max 200 words)</small>
                            </label>
                            <textarea
                                id="improvementSuggestions"
                                className="form-control"
                                rows={3}
                                value={improvementSuggestions}
                                onChange={(e) => setImprovementSuggestions(e.target.value)}
                                placeholder="What was confusing? What could be improved?"
                                maxLength={200}
                            />
                            <small className="form-text text-muted">
                                {improvementSuggestions.length}/200 characters
                            </small>
                        </div>

                        <div className="mb-3">
                            <label htmlFor="timeSpent" className="form-label">
                                How much time did you spend on this assignment? <small className="text-muted">(optional)</small>
                            </label>
                            <input
                                id="timeSpent"
                                type="text"
                                className="form-control"
                                value={timeSpent}
                                onChange={(e) => setTimeSpent(e.target.value)}
                                placeholder="e.g., 2 hours, 3 days, 1 week"
                                maxLength={100}
                            />
                        </div>

                        <div className="mb-3 form-check">
                            <input
                                id="anonymous"
                                type="checkbox"
                                className="form-check-input"
                                checked={anonymous}
                                onChange={(e) => setAnonymous(e.target.checked)}
                            />
                            <label htmlFor="anonymous" className="form-check-label">
                                Submit feedback anonymously
                            </label>
                        </div>

                        <div className="d-flex gap-2">
                            <button
                                type="submit"
                                className="btn btn-primary"
                                disabled={isSubmitting || (likedContent.trim().length < 10 && improvementSuggestions.trim().length < 10)}
                            >
                                {isSubmitting ? (
                                    <>
                                        <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true" />
                                        Submitting...
                                    </>
                                ) : (
                                    'Submit Feedback'
                                )}
                            </button>
                            <button
                                type="button"
                                className="btn btn-secondary"
                                onClick={() => setIsOpen(false)}
                            >
                                Cancel
                            </button>
                        </div>
                    </form>
                </div>
            )}
        </div>
    )
}

export default AssignmentFeedbackForm
