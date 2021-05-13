import React, { useEffect, useState } from "react"
import { Link, useHistory } from "react-router-dom"
import { useOvermind } from "../overmind"
import { Enrollment, Submission } from "../../proto/ag_pb"
import { ProgressBar } from "./ProgressBar"


const NavBarLabs = () => {
    const {state} = useOvermind()
    
    const history  = useHistory()
    
    const redirectToLab = (assignmentid: number) => {
        history.push(`/course/${state.activeCourse}/${assignmentid}`)
    }

    const Links: Function = (): JSX.Element[] => { 
        
        if(state.assignments[state.activeCourse]) {
            let links = state.assignments[state.activeCourse]?.map((assignment, index) => {
                // Class name to determine background color
                let active = state.activeLab === assignment.getId() && state.activeCourse === assignment.getCourseid() ? "active" : ""

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