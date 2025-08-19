import React from 'react';
const SubmissionTypeIcon = ({ solo }) => {
    const indicator = solo ? "fa-user" : "fa-users";
    return (React.createElement("i", { className: `fa ${indicator} submission-icon`, title: solo ? "Solo Submission" : "Group Submission" }));
};
export default SubmissionTypeIcon;
