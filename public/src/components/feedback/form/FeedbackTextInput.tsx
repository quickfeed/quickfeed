import React from 'react'

interface FeedbackTextInputProps {
    id: string
    label: string
    value: string
    onChange: (value: string) => void
    placeholder: string
    wordCount: number
    maxWords: number
    minWords: number
}

export const FeedbackTextInput: React.FC<FeedbackTextInputProps> = ({
    id,
    label,
    value,
    onChange,
    placeholder,
    wordCount,
    maxWords,
    minWords
}) => {
    return (
        <div className="mb-3">
            <label htmlFor={id} className="form-label">
                {label} <small className="text-muted">(min {minWords} words, max {maxWords} words)</small>
            </label>
            <textarea
                id={id}
                className="form-control"
                rows={3}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                placeholder={placeholder}
                maxLength={2000}
            />
            <small className="form-text text-muted">
                {wordCount}/{maxWords} words
            </small>
        </div>
    )
}

export default FeedbackTextInput
