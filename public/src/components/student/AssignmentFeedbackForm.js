import React, { useState } from 'react';
import { useActions } from '../../overmind';
const AssignmentFeedbackForm = ({ assignment, courseID }) => {
    const actions = useActions();
    const [isOpen, setIsOpen] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isSubmitted, setIsSubmitted] = useState(false);
    const [likedContent, setLikedContent] = useState('');
    const [improvementSuggestions, setImprovementSuggestions] = useState('');
    const [timeSpent, setTimeSpent] = useState('');
    const [anonymous, setAnonymous] = useState(true);
    const handleSubmit = async (e) => {
        e.preventDefault();
        if (likedContent.trim().length < 10 && improvementSuggestions.trim().length < 10) {
            alert('Please provide at least 10 words in either "What did you like?" or "What would make it better?"');
            return;
        }
        if (likedContent.length > 200 || improvementSuggestions.length > 200 || timeSpent.length > 100) {
            alert('Please keep responses under the word limit (200 words for feedback, 100 for time spent)');
            return;
        }
        setIsSubmitting(true);
        try {
            const feedback = {
                assignmentID: assignment.ID,
                userID: anonymous ? BigInt(0) : undefined,
                likedContent: likedContent.trim(),
                improvementSuggestions: improvementSuggestions.trim(),
                timeSpent: timeSpent.trim(),
                commitHash: '',
                submissionID: BigInt(0),
                createdAt: undefined,
                ID: BigInt(0)
            };
            await actions.feedback.createAssignmentFeedback({ courseID, feedback });
            setIsSubmitted(true);
            setIsOpen(false);
            setLikedContent('');
            setImprovementSuggestions('');
            setTimeSpent('');
        }
        catch (error) {
            console.error('Failed to submit feedback:', error);
            alert('Failed to submit feedback. Please try again.');
        }
        finally {
            setIsSubmitting(false);
        }
    };
    if (isSubmitted) {
        return (React.createElement("div", { className: "card mt-3" },
            React.createElement("div", { className: "card-body" },
                React.createElement("h5", { className: "card-title text-success" },
                    React.createElement("i", { className: "fa fa-check-circle me-2" }),
                    "Feedback Submitted"),
                React.createElement("p", { className: "card-text" },
                    "Thank you for your feedback on \"",
                    assignment.name,
                    "\"!"))));
    }
    return (React.createElement("div", { className: "card mt-3" },
        React.createElement("div", { className: "card-header" },
            React.createElement("button", { className: "btn btn-link p-0 text-decoration-none w-100 text-start", onClick: () => setIsOpen(!isOpen), type: "button", "aria-expanded": isOpen },
                React.createElement("h5", { className: "mb-0" },
                    React.createElement("i", { className: `fa fa-chevron-${isOpen ? 'down' : 'right'} me-2` }),
                    "Give Feedback on This Assignment"))),
        isOpen && (React.createElement("div", { className: "card-body" },
            React.createElement("form", { onSubmit: handleSubmit },
                React.createElement("div", { className: "mb-3" },
                    React.createElement("label", { htmlFor: "likedContent", className: "form-label" },
                        "What did you like about this assignment? ",
                        React.createElement("small", { className: "text-muted" }, "(min 10 words, max 200 words)")),
                    React.createElement("textarea", { id: "likedContent", className: "form-control", rows: 3, value: likedContent, onChange: (e) => setLikedContent(e.target.value), placeholder: "What worked well? What was interesting or helpful?", maxLength: 200 }),
                    React.createElement("small", { className: "form-text text-muted" },
                        likedContent.length,
                        "/200 characters")),
                React.createElement("div", { className: "mb-3" },
                    React.createElement("label", { htmlFor: "improvementSuggestions", className: "form-label" },
                        "What would make this assignment better? ",
                        React.createElement("small", { className: "text-muted" }, "(min 10 words, max 200 words)")),
                    React.createElement("textarea", { id: "improvementSuggestions", className: "form-control", rows: 3, value: improvementSuggestions, onChange: (e) => setImprovementSuggestions(e.target.value), placeholder: "What was confusing? What could be improved?", maxLength: 200 }),
                    React.createElement("small", { className: "form-text text-muted" },
                        improvementSuggestions.length,
                        "/200 characters")),
                React.createElement("div", { className: "mb-3" },
                    React.createElement("label", { htmlFor: "timeSpent", className: "form-label" },
                        "How much time did you spend on this assignment? ",
                        React.createElement("small", { className: "text-muted" }, "(optional)")),
                    React.createElement("input", { id: "timeSpent", type: "text", className: "form-control", value: timeSpent, onChange: (e) => setTimeSpent(e.target.value), placeholder: "e.g., 2 hours, 3 days, 1 week", maxLength: 100 })),
                React.createElement("div", { className: "mb-3 form-check" },
                    React.createElement("input", { id: "anonymous", type: "checkbox", className: "form-check-input", checked: anonymous, onChange: (e) => setAnonymous(e.target.checked) }),
                    React.createElement("label", { htmlFor: "anonymous", className: "form-check-label" }, "Submit feedback anonymously")),
                React.createElement("div", { className: "d-flex gap-2" },
                    React.createElement("button", { type: "submit", className: "btn btn-primary", disabled: isSubmitting || (likedContent.trim().length < 10 && improvementSuggestions.trim().length < 10) }, isSubmitting ? (React.createElement(React.Fragment, null,
                        React.createElement("span", { className: "spinner-border spinner-border-sm me-2", role: "status", "aria-hidden": "true" }),
                        "Submitting...")) : ('Submit Feedback')),
                    React.createElement("button", { type: "button", className: "btn btn-secondary", onClick: () => setIsOpen(false) }, "Cancel")))))));
};
export default AssignmentFeedbackForm;
