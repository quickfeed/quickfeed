import React, { useEffect } from "react"
import { Link, useHistory } from "react-router-dom"
import { useAppState } from "../../overmind"
import { Submission } from "../../../proto/ag/ag_pb"
import { ProgressBar } from "../ProgressBar"


const NavBarLabs = () => {
    const state = useAppState()
    const history  = useHistory()
    
    const redirectToLab = (assignmentID: number) => {
        history.push(`/course/${state.activeCourse}/${assignmentID}`)
    }

    const Links: Function = (): JSX.Element[] => { 
        if(state.assignments[state.activeCourse] && state.submissions[state.activeCourse]) {
            const links = state.assignments[state.activeCourse]?.map((assignment, index) => {
                // Class name to determine background color
                const active = state.activeLab === assignment.getId() && state.activeCourse === assignment.getCourseid() ? "active" : ""

                return (
                    <li style={{position: "relative", height: "50px"}} className={active} key={assignment.getId()} onClick={() => {redirectToLab(assignment.getId())}}>
                        <div id="icon">
                            <i className={assignment.getIsgrouplab() ? "fa fa-users" : "fa fa-user"} title={assignment.getIsgrouplab() ? "Group Lab" : "Individual Lab"}>
                            </i>
                        </div>
                        <div id="title">
                            <Link to={`/course/${state.activeCourse}/${assignment.getId()}`}>
                                {assignment.getName()}
                            </Link>
                        </div> 
                        {state.submissions[state.activeCourse][assignment.getOrder() - 1]?.getStatus() === Submission.Status.APPROVED && 
                            <i className="fa fa-check" style={{marginLeft: "10px"}}></i>
                        }
                        <ProgressBar courseID={state.activeCourse} assignmentIndex={index} type="navbar" />

                    </li>
                )
            })
            return links
        }
        return []

    }

    // Render
    return (
        <React.Fragment>
            <Links />
        </React.Fragment>
    )
}

export default NavBarLabs