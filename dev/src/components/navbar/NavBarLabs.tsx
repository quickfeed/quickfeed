import React from "react"
import { useHistory } from "react-router-dom"
import { useAppState } from "../../overmind"
import { Assignment, Submission } from "../../../proto/ag/ag_pb"
import { Progress, ProgressBar } from "../ProgressBar"
import NavBarLink, { NavLink } from "./NavBarLink"


const NavBarLabs = (): JSX.Element => {
    const state = useAppState()
    const history  = useHistory()
    
    if (!state.assignments[state.activeCourse] || !state.submissions[state.activeCourse]) {
        return
    }

    const redirectToLab = (assignmentID: number) => {
        history.push(`/course/${state.activeCourse}/${assignmentID}`)
    }

    const submissionIcon = (assignment: Assignment) => {
        const submission = state.submissions[state.activeCourse][assignment.getOrder() - 1]
        return (
            <div>
                <i className={assignment.getIsgrouplab() ? "fa fa-users" : "fa fa-user"} title={assignment.getIsgrouplab() ? "Group Lab" : "Individual Lab"} />
                {submission?.getStatus() === Submission.Status.APPROVED && <i className="fa fa-check" style={{marginLeft: "10px"}} />}
            </div>
        )
    }

    const getLinkClass = (assignment: Assignment) => {
        return state.activeLab === assignment.getId() && state.activeCourse === assignment.getCourseid() ? "active" : ""
    }

    const labLinks = state.assignments[state.activeCourse]?.map((assignment, index) => {
        const link: NavLink = {link: {text: assignment.getName(), to: `/course/${state.activeCourse}/${assignment.getId()}`}, jsx: submissionIcon(assignment)}
        return (
            <div className={getLinkClass(assignment)} style={{position: "relative"}} key={assignment.getId()} onClick={() => {redirectToLab(assignment.getId())}}>
                <NavBarLink link={link.link} jsx={link.jsx}/>
                <ProgressBar courseID={state.activeCourse} assignmentIndex={index} type={Progress.NAV} />
            </div>
        )
    })
    

    return (
        <>
            {labLinks}
        </>
    )
}

export default NavBarLabs