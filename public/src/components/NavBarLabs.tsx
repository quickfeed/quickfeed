import React, { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"
import { Submission } from "../proto/ag_pb"
import { ProgressBar } from "./ProgressBar"


const NavBarLabs = () => {
    const {state} = useOvermind()

    const [active, setActive] = useState(0)
    
    useEffect(() => {

    })

    const labs = (): JSX.Element[] => { 
        
        if(state.assignments[state.activeCourse] !== undefined && state.activeCourse > 0) {
            let links = state.assignments[state.activeCourse]?.map((assignment, index) => {
                return (
                    <li style={{position: "relative"}} className={state.activeLab === assignment.getId() && state.activeCourse === assignment.getCourseid() ? "active" : ""} key={assignment.getId()} onClick={() => {setActive(assignment.getId())}}>
                        <div id="icon">
                            <i className={assignment.getIsgrouplab() ? "fa fa-users" : "fa fa-user"} title={assignment.getIsgrouplab() ? "Group Lab" : "Individual Lab"}>
                            </i>
                        </div>
                        <div id="title">
                            <Link to={`/course/${state.activeCourse}/${assignment.getId()}`}>{assignment.getName()}</Link>
                        </div> 
                        
                        <ProgressBar courseID={state.activeCourse} assignmentID={index} type="navbar" />

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
            {labs()}
        </React.Fragment>
    )
}

export default NavBarLabs