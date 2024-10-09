import React from "react"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import NavBarLink, { NavLink } from "./NavBarLink"
import { useHistory } from "react-router"
import { Status } from "../../consts"
import { getStatusByUser, isApproved, isGroupSubmission, isValidSubmissionForAssignment } from "../../Helpers"
import SubmissionTypeIcon from "../student/SubmissionTypeIcon"


const NavBarLabs = (): JSX.Element | null => {
    const state = useAppState()
    const history = useHistory()

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

    const getLinkClass = (assignment: Assignment) => {
        return BigInt(state.selectedAssignmentID) === assignment.ID ? Status.Active : ""
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
            const link: NavLink = { link: { text: assignment.name, to: `/course/${state.activeCourse}/${isGroupSubmission(submission) ? "group-lab": "lab"}/${assignment.ID}` }, jsx: submissionIcon(submission) }
            return (
                <div className={getLinkClass(assignment)} style={{ position: "relative" }} key={assignment.ID.toString()} onClick={() => { history.push(link.link.to) }}>
                    <NavBarLink link={link.link} jsx={link.jsx} />
                    <ProgressBar courseID={state.activeCourse.toString()} submission={submission} type={Progress.NAV} />
                </div>
            )
        })
    })

    return <>{labLinks}</>
}

export default NavBarLabs
