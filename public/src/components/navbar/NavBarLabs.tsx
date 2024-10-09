import React from "react"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import NavBarLink, { NavLink } from "./NavBarLink"
import { useHistory } from "react-router"
import { Status } from "../../consts"
import { getStatusByUser, isApproved } from "../../Helpers"


const NavBarLabs = (): JSX.Element | null => {
    const state = useAppState()
    const history = useHistory()

    if (!state.assignments[state.activeCourse.toString()]) {
        return null
    }

    const submissionIcon = (assignment: Assignment, submission: Submission) => {
        const icon: JSX.Element | null = null
        if (assignment.isGroupLab) {
            if (submission.groupID !== 0n) {
                return <i className={"fa fa-users"} title={"Group assignment"} />
            }
            if (submission.userID !== 0n) {
                return <i className={"fa fa-user"} title={"Solo submission"} />
            }
        }
        return (
            <div>
                {icon}
                {isApproved(getStatusByUser(submission, state.self.ID)) && <i className="fa fa-check ml-2" />}
            </div>
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
            const groupSubmission = submission.groupID !== 0n
            if (!assignment.isGroupLab && groupSubmission) {
                return null
            }
            const link: NavLink = { link: { text: assignment.name, to: `/course/${state.activeCourse}/${groupSubmission ? "group-lab": "lab"}/${assignment.ID}` }, jsx: submissionIcon(assignment, submission) }
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
