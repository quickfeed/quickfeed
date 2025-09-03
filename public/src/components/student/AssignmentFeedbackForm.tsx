import React, { useState } from 'react'
import { Assignment, AssignmentFeedback, AssignmentFeedbackSchema } from '../../../proto/qf/types_pb'
import { useActions, useAppState } from '../../overmind'
import { create } from "@bufbuild/protobuf"
import { Color } from "../../Helpers"

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

        // Basic validation
        if (likedContent.trim().length < 10 && improvementSuggestions.trim().length < 10) {
            actions.global.alert({ color: Color.RED, text: 'Please provide at least 10 words in either "What did you like?" or "What would make it better?"' })
            return
        }

        if (likedContent.length > 200 || improvementSuggestions.length > 200 || (timeSpent && Number(timeSpent) > 200)) {
            actions.global.alert({ color: Color.RED, text: 'Please keep responses under the word limit (200 words for feedback, 200 for time spent)' })
            return
        }

        setIsSubmitting(true)


        const feedback: AssignmentFeedback = create(AssignmentFeedbackSchema, {
            ID: BigInt(0), // Will be set by backend
            CourseID: assignment.CourseID,
            AssignmentID: assignment.ID,
            UserID: anonymous ? BigInt(0) : state.self.ID, // Will be set by backend if not anonymous
            LikedContent: likedContent.trim(),
            ImprovementSuggestions: improvementSuggestions.trim(),
            TimeSpent: timeSpent ? Number(timeSpent) : 0,
            CommitHash: '', // Could be populated from current submission
            SubmissionID: BigInt(0), // Could be populated from current submission
            CreatedAt: undefined, // Will be set by backend
        })

        const success = await actions.feedback.createAssignmentFeedback({ courseID, feedback })
        if (!success) {
            actions.global.alert({ color: Color.RED, text: 'Failed to submit feedback. Please try again later.' })
            setIsSubmitting(false)
            return
        }
        setIsSubmitted(true)
        setIsOpen(false)

        // Reset form
        setLikedContent('')
        setImprovementSuggestions('')
        setTimeSpent('')


        setIsSubmitting(false)

    }

    if (isSubmitted) {
        return (
            <div className="card mt-3">
                <div className="card-body">
                    <h5 className="card-title text-success">
                        <i className="fa fa-check-circle me-2"></i>
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
                        <i className={`fa fa-chevron-${isOpen ? 'down' : 'right'} me-2`}></i>
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
                                type="number"
                                className="form-control"
                                value={timeSpent}
                                onChange={(e) => setTimeSpent(e.target.value)}
                                placeholder="How much time did you spend on this assignment? (in hours)"
                                max={200}
                                min={0}
                            />
                        </div>


                        <div className="d-flex justify-content-end">
                            <div className="form-check align-self-center">
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
                            <button
                                type="submit"
                                className="btn btn-primary ml-2"
                                disabled={isSubmitting || (likedContent.trim().length < 10 && improvementSuggestions.trim().length < 10)}
                            >
                                {isSubmitting ? (
                                    <>
                                        <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                                        Submitting...
                                    </>
                                ) : (
                                    'Submit Feedback'
                                )}
                            </button>
                            <button
                                type="button"
                                className="btn btn-secondary ml-2"
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
