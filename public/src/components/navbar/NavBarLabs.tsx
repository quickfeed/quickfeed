import React from "react"
import { useAppState } from "../../overmind"
import { Assignment, Submission_Status } from "../../../proto/qf/types_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import NavBarLink, { NavLink } from "./NavBarLink"
import { useHistory } from "react-router"
import { Status } from "../../consts"


const NavBarLabs = (): JSX.Element | null => {
    const state = useAppState()
    const history = useHistory()

    if (!state.assignments[state.activeCourse.toString()] || !state.submissions[state.activeCourse.toString()]) {
        return null
    }

    const submissionIcon = (assignment: Assignment) => {
        const submission = state.submissions[state.activeCourse.toString()][assignment.order - 1]
        return (
            <div>
                {assignment.isGroupLab && <i className={"fa fa-users"} title={"Group assignment"} />}
                {submission?.status === Submission_Status.APPROVED && <i className="fa fa-check ml-2" />}
            </div>
        )
    }

    const getLinkClass = (assignment: Assignment) => {
        return BigInt(state.activeAssignment) === assignment.ID ? Status.Active : ""
    }

    const labLinks = state.assignments[state.activeCourse.toString()]?.map((assignment, index) => {
        const link: NavLink = { link: { text: assignment.name, to: `/course/${state.activeCourse}/lab/${assignment.ID}` }, jsx: submissionIcon(assignment) }
        return (
            <div className={getLinkClass(assignment)} style={{ position: "relative" }} key={assignment.ID.toString()} onClick={() => { history.push(link.link.to) }}>
                <NavBarLink link={link.link} jsx={link.jsx} />
                <ProgressBar courseID={state.activeCourse.toString()} assignmentIndex={index} type={Progress.NAV} />
            </div>
        )
    })

    return <>{labLinks}</>
}

export default NavBarLabs
