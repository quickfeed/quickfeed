import React from "react"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/qf/qf_pb"
import ProgressBar, { Progress } from "../ProgressBar"
import NavBarLink, { NavLink } from "./NavBarLink"
import { useHistory } from "react-router"
import { Status } from "../../consts"


const NavBarLabs = (): JSX.Element | null => {
    const state = useAppState()
    const history = useHistory()

    if (!state.assignments[state.activeCourse] || !state.submissions[state.activeCourse]) {
        return null
    }

    const submissionIcon = (assignment: Assignment.AsObject) => {
        const submission = state.submissions[state.activeCourse][assignment.order - 1]
        return (
            <div>
                {assignment.isgrouplab && <i className={"fa fa-users"} title={"Group assignment"} />}
                {submission?.status === Submission.Status.APPROVED && <i className="fa fa-check ml-2" />}
            </div>
        )
    }

    const getLinkClass = (assignment: Assignment.AsObject) => {
        return state.activeAssignment === assignment.id ? Status.Active : ""
    }

    const labLinks = state.assignments[state.activeCourse]?.map((assignment, index) => {
        const link: NavLink = { link: { text: assignment.name, to: `/course/${state.activeCourse}/${assignment.id}` }, jsx: submissionIcon(assignment) }
        return (
            <div className={getLinkClass(assignment)} style={{ position: "relative" }} key={assignment.id} onClick={() => { history.push(link.link.to) }}>
                <NavBarLink link={link.link} jsx={link.jsx} />
                <ProgressBar courseID={state.activeCourse} assignmentIndex={index} type={Progress.NAV} />
            </div>
        )
    })

    return <>{labLinks}</>
}

export default NavBarLabs
