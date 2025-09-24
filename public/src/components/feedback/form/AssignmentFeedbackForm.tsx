import { create } from "@bufbuild/protobuf"
import React, { useState } from 'react'
import { Assignment, AssignmentFeedback, AssignmentFeedbackSchema } from '../../../../proto/qf/types_pb'
import { Color } from "../../../Helpers"
import { useActions } from '../../../overmind'
import FeedbackSubmittedCard from "./FeedbackSubmitted"
import FeedbackTextInput from "./FeedbackTextInput"
import FeedbackFormActions from "./FeedbackFormActions"
import TimeSpentInput from "./TimeSpentInput"

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

    const minWords = 1

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

    const validateForm = (): boolean => {
        const likedWordsCount = countWords(likedContent)
        const improvementWordsCount = countWords(improvementSuggestions)

        return (likedWordsCount >= minWords || improvementWordsCount >= minWords) &&
            likedWordsCount <= 200 &&
            improvementWordsCount <= 200 &&
            timeSpent > 0 &&
            timeSpent <= 6000
    }

    const countWords = (text: string): number => {
        return text.trim().split(/\s+/).filter(word => word.length > 0).length
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        const likedWordsCount = countWords(likedContent)
        const improvementWordsCount = countWords(improvementSuggestions)

        if (likedWordsCount < minWords && improvementWordsCount < minWords) {
            actions.global.alert({ color: Color.RED, text: `Please provide at least ${minWords} words in either "What did you like?" or "What would make it better?"` })
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
        return <FeedbackSubmittedCard assignment={assignment} />
    }

    return (
        <div className="card mt-3 mb-3">
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
                        <FeedbackTextInput
                            id="likedContent"
                            label="What did you like about this assignment?"
                            value={likedContent}
                            onChange={setLikedContent}
                            placeholder="What worked well? What was interesting or helpful?"
                            wordCount={countWords(likedContent)}
                            maxWords={200}
                            minWords={minWords}
                        />
                        <FeedbackTextInput
                            id="improvementSuggestions"
                            label="What would make this assignment better?"
                            value={improvementSuggestions}
                            onChange={setImprovementSuggestions}
                            placeholder="What was confusing? What could be improved?"
                            wordCount={countWords(improvementSuggestions)}
                            maxWords={200}
                            minWords={minWords}
                        />
                        <TimeSpentInput
                            hours={hours}
                            minutes={minutes}
                            onHoursChange={handleHoursChange}
                            onMinutesChange={handleMinutesChange}
                        />
                        <div className="mb-3">
                            <small className="text-muted">
                                <i className="fa fa-info-circle me-1" />
                                Your feedback will be submitted anonymously to help improve the course.
                            </small>
                        </div>
                        <FeedbackFormActions
                            isSubmitting={isSubmitting}
                            isFormValid={validateForm()}
                            onCancel={() => setIsOpen(false)}
                        />
                    </form >
                </div >
            )}
        </div >
    )
}

export default AssignmentFeedbackForm
