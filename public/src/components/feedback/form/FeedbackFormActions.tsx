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
        <div className="d-flex justify-content-end">
            <button
                type="submit"
                className="btn btn-primary ml-2"
                disabled={isSubmitting || !isFormValid}
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
                onClick={onCancel}
            >
                Cancel
            </button>
        </div>
    )
}

export default FeedbackFormActions
