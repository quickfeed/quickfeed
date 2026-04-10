import React from 'react'

interface FeedbackFormActionsProps {
    isSubmitting: boolean
    isFormValid: boolean
    onCancel: () => void
}

export const FeedbackFormActions: React.FC<FeedbackFormActionsProps> = ({
    isSubmitting,
    isFormValid,
    onCancel
}) => {
    return (
        <div className="flex flex-col sm:flex-row justify-end gap-3 pt-4 border-t border-base-300">
            <button
                type="button"
                className="btn btn-ghost"
                onClick={onCancel}
            >
                Cancel
            </button>
            <button
                type="submit"
                className="btn btn-primary gap-2"
                disabled={isSubmitting || !isFormValid}
            >
                {isSubmitting ? (
                    <>
                        <span className="loading loading-spinner loading-sm" />
                        Submitting...
                    </>
                ) : (
                    <>
                        <i className="fa fa-paper-plane" />
                        Submit Feedback
                    </>
                )}
            </button>
        </div>
    )
}

export default FeedbackFormActions
