import React from "react"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import NavBarLink, { NavLink } from "./NavBarLink"
import { useNavigate, useLocation } from "react-router"
import { Status } from "../../consts"
import { getStatusByUser, isApproved, isGroupSubmission, isValidSubmissionForAssignment } from "../../Helpers"
import SubmissionTypeIcon from "../student/SubmissionTypeIcon"


const NavBarLabs = () => {
    const state = useAppState()
    const navigate = useNavigate()
    const location = useLocation()

    if (!state.assignments[state.activeCourse.toString()]) {
        return null
    }

    const submissionIcon = (submission: Submission) => {
        return (
            <>
                <SubmissionTypeIcon solo={!isGroupSubmission(submission)} />
                {isApproved(getStatusByUser(submission, state.self.ID)) && <i className="fa fa-check ml-2" />}
            </>
        )
    }

    const highlightSubmission = (submission: Submission, assignment: Assignment) => {
        // The submission should be highlighted if:
        // - the assignment ID is equal to the selected assignment ID
        //  AND ONE OF THE FOLLOWING:
        // - the location contains `group-lab` and the submission is a group submission
        // - the location contains `lab` and the submission is not a group submission
        // Otherwise, return an empty string
        // This way we can highlight the correct lab link in the navbar
        let linkClass = ""
        if (BigInt(state.selectedAssignmentID) === assignment.ID) {
            const groupPath = location.pathname.includes("group-lab")
            if (groupPath && isGroupSubmission(submission)) {
                linkClass = Status.Active
            } else if (!groupPath && !isGroupSubmission(submission)) {
                linkClass = Status.Active
            }
        }
        return linkClass
    }

    const labLinks = state.assignments[state.activeCourse.toString()]?.map(assignment => {
        const submissions = state.submissions.ForAssignment(assignment)
        if (!submissions) {
            return null
        }
        return submissions.map(submission => {
            if (!isValidSubmissionForAssignment(submission, assignment)) {
                return null
            }
            const link: NavLink = {
                text: assignment.name,
                to: `/course/${state.activeCourse}/${isGroupSubmission(submission) ? "group-lab" : "lab"}/${assignment.ID}`,
                jsx: submissionIcon(submission)
            }
            return (
                <div
                    className={highlightSubmission(submission, assignment)}
                    style={{ position: "relative" }}
                    key={submission.ID.toString()}
                    onClick={() => { navigate(link.to) }}
                    role="button"
                    aria-hidden="true"
                >
                    <NavBarLink link={link} />
                    <ProgressBar courseID={state.activeCourse.toString()} submission={submission} type={Progress.NAV} />
                </div>
            )
        })
    })

    return <>{labLinks}</>
}

export default NavBarLabs
