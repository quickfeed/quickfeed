import React from "react";
import { useAppState } from "../../overmind";
import ProgressBar, { Progress } from "../ProgressBar";
import NavBarLink from "./NavBarLink";
import { useNavigate, useLocation } from "react-router";
import { Status } from "../../consts";
import { getStatusByUser, isApproved, isGroupSubmission, isValidSubmissionForAssignment } from "../../Helpers";
import SubmissionTypeIcon from "../student/SubmissionTypeIcon";
const NavBarLabs = () => {
    const state = useAppState();
    const navigate = useNavigate();
    const location = useLocation();
    if (!state.assignments[state.activeCourse.toString()]) {
        return null;
    }
    const submissionIcon = (submission) => {
        return (React.createElement(React.Fragment, null,
            React.createElement(SubmissionTypeIcon, { solo: !isGroupSubmission(submission) }),
            isApproved(getStatusByUser(submission, state.self.ID)) && React.createElement("i", { className: "fa fa-check ml-2" })));
    };
    const highlightSubmission = (submission, assignment) => {
        let linkClass = "";
        if (BigInt(state.selectedAssignmentID) === assignment.ID) {
            const groupPath = location.pathname.includes("group-lab");
            if (groupPath && isGroupSubmission(submission)) {
                linkClass = Status.Active;
            }
            else if (!groupPath && !isGroupSubmission(submission)) {
                linkClass = Status.Active;
            }
        }
        return linkClass;
    };
    const labLinks = state.assignments[state.activeCourse.toString()]?.map(assignment => {
        const submissions = state.submissions.ForAssignment(assignment);
        if (!submissions) {
            return null;
        }
        return submissions.map(submission => {
            if (!isValidSubmissionForAssignment(submission, assignment)) {
                return null;
            }
            const link = {
                text: assignment.name,
                to: `/course/${state.activeCourse}/${isGroupSubmission(submission) ? "group-lab" : "lab"}/${assignment.ID}`,
                jsx: submissionIcon(submission)
            };
            return (React.createElement("div", { className: highlightSubmission(submission, assignment), style: { position: "relative" }, key: submission.ID.toString(), onClick: () => { navigate(link.to); }, role: "button", "aria-hidden": "true" },
                React.createElement(NavBarLink, { link: link }),
                React.createElement(ProgressBar, { courseID: state.activeCourse.toString(), submission: submission, type: Progress.NAV })));
        });
    });
    return React.createElement(React.Fragment, null, labLinks);
};
export default NavBarLabs;
