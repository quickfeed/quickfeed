import { create } from "@bufbuild/protobuf"
import React, { useState } from 'react'
import { Assignment, AssignmentFeedback, AssignmentFeedbackSchema } from '../../../proto/qf/types_pb'
import { Color } from "../../Helpers"
import { useActions } from '../../overmind'

interface AssignmentFeedbackFormProps {
    assignment: Assignment
    courseID: string
}

const AssignmentFeedbackForm: React.FC<AssignmentFeedbackFormProps> = ({ assignment, courseID }) => {
    const actions = useActions()
    const [isOpen, setIsOpen] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [isSubmitted, setIsSubmitted] = useState(false)
    const [likedContent, setLikedContent] = useState('')
    const [improvementSuggestions, setImprovementSuggestions] = useState('')
    const [timeSpent, setTimeSpent] = useState(0) // in hours
    const [hours, setHours] = useState('')
    const [minutes, setMinutes] = useState('')

    const validateTimeInput = (value: string, max: number): boolean => {
        if (value === '') return true
        const num = parseInt(value, 10)
        return !isNaN(num) && num >= 0 && num <= max
    }

    const handleHoursChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value
        if (validateTimeInput(value, 100)) {
            setHours(value)
            // timeSpent is a combination of hours and minutes in minutes
            const totalHours = parseInt(value, 10)
            const totalMinutes = parseInt(minutes || '0', 10)
            setTimeSpent(totalHours * 60 + totalMinutes)
        }
    }

    const handleMinutesChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value
        if (validateTimeInput(value, 59)) {
            setMinutes(value)
            // timeSpent is a combination of hours and minutes in minutes
            const totalHours = parseInt(hours || '0', 10)
            const totalMinutes = parseInt(value, 10)
            setTimeSpent(totalHours * 60 + totalMinutes)
        }
    }

    const countWords = (text: string): number => {
        return text.trim().split(/\s+/).filter(word => word.length > 0).length
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        const likedWordsCount = countWords(likedContent)
        const improvementWordsCount = countWords(improvementSuggestions)

        if (likedWordsCount < 10 && improvementWordsCount < 10) {
            actions.global.alert({ color: Color.RED, text: 'Please provide at least 10 words in either "What did you like?" or "What would make it better?"' })
            return
        }

        if (likedWordsCount > 200 || improvementWordsCount > 200 || (timeSpent && Number(timeSpent) > 6000)) {
            actions.global.alert({ color: Color.RED, text: 'Please keep responses under the word limit (200 words for feedback, 100 hours for time spent)' })
            return
        }

        // Validate time input
        if (!hours && !minutes) {
            actions.global.alert({ text: 'Please specify the time you spent on this assignment', color: Color.YELLOW })
            return
        }

        setIsSubmitting(true)


        const feedback: AssignmentFeedback = create(AssignmentFeedbackSchema, {
            ID: BigInt(0), // Will be set by backend
            CourseID: assignment.CourseID,
            AssignmentID: assignment.ID,
            LikedContent: likedContent.trim(),
            ImprovementSuggestions: improvementSuggestions.trim(),
            TimeSpent: timeSpent ? Number(timeSpent) : 0,
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
        setTimeSpent(0)


        setIsSubmitting(false)

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
                                maxLength={2000}
                            />
                            <small className="form-text text-muted">
                                {countWords(likedContent)}/200 words
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
                                maxLength={2000}
                            />
                            <small className="form-text text-muted">
                                {countWords(improvementSuggestions)}/200 words
                            </small>
                        </div>

                        <div className="mb-3">
                            <label className="form-label">
                                How much time did you spend on this assignment? <span className="text-danger">*</span>
                            </label>
                            <p className="text-muted small mb-2">Enter hours and minutes (max 100 hours)</p>
                            <div className="row">
                                <div className="col-6">
                                    <div className="input-group">
                                        <input
                                            type="number"
                                            className="form-control"
                                            value={hours}
                                            onChange={handleHoursChange}
                                            placeholder="0"
                                            min="0"
                                            max="100"
                                        />
                                        <span className="input-group-text">hours</span>
                                    </div>
                                </div>
                                <div className="col-6">
                                    <div className="input-group">
                                        <input
                                            type="number"
                                            className="form-control"
                                            value={minutes}
                                            onChange={handleMinutesChange}
                                            placeholder="0"
                                            min="0"
                                            max="59"
                                        />
                                        <span className="input-group-text">minutes</span>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="mb-3">
                            <small className="text-muted">
                                <i className="fa fa-info-circle me-1"></i>
                                Your feedback will be submitted anonymously to help improve the course.
                            </small>
                        </div>

                        <div className="d-flex justify-content-end">
                            <button
                                type="submit"
                                className="btn btn-primary ml-2"
                                disabled={
                                    isSubmitting ||
                                    (
                                        countWords(likedContent) < 10 ||
                                        countWords(improvementSuggestions) < 10 ||
                                        timeSpent > 6000 || timeSpent === 0
                                    )
                                }
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
                                className="btn btn-secondary ml-2"
                                onClick={() => setIsOpen(false)}
                            >
                                Cancel
                            </button>
                        </div>
                    </form >
                </div >
            )}
        </div >
    )
}

export default AssignmentFeedbackForm
