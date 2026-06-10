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
        <div className="w-full">
            <div className="mb-2">
                <label htmlFor={id} className="block text-base font-semibold text-base-content mb-1">
                    {label}
                </label>
                <div className="text-xs text-base-content/60">
                    min {minWords}, max {maxWords} words
                </div>
            </div>
            <textarea
                id={id}
                className="textarea h-24 textarea-primary w-full"
                value={value}
                onChange={(e) => onChange(e.target.value)}
                placeholder={placeholder}
                maxLength={2000}
            />
            <div className="flex justify-end mt-1">
                <span className={`text-xs font-semibold ${wordCount > maxWords ? 'text-error' : wordCount >= minWords ? 'text-success' : 'text-base-content/60'
                    }`}>
                    {wordCount}/{maxWords} words
                </span>
            </div>
        </div>
    )
}

export default FeedbackTextInput
