import React from 'react';
export var KnownMessage;
(function (KnownMessage) {
    KnownMessage["TeacherNoSubmission"] = "Select a submission from the results table";
    KnownMessage["TeacherNoAssignment"] = "Assignment does not have a submission";
    KnownMessage["StudentNoSubmission"] = "No submission found";
    KnownMessage["StudentNoAssignment"] = "Assignment not found";
})(KnownMessage || (KnownMessage = {}));
export const CenteredMessage = ({ message }) => {
    return React.createElement("div", { className: "text-center mt-5" },
        React.createElement("h3", null, message));
};
