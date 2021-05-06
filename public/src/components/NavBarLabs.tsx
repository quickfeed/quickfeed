import React, { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"
import { Submission } from "../../proto/ag_pb"
import { ProgressBar } from "./ProgressBar"


const NavBarLabs = () => {
    const {state} = useOvermind()

    const [active, setActive] = useState(-1)
    
    useEffect(() => {
    }, [state.activeCourse])

    const labs = (): JSX.Element[] => { 
        
        if(state.assignments[state.activeCourse] !== undefined && state.activeCourse > 0) {
            let links = state.assignments[state.activeCourse]?.map((assignment, index) => {
                let rand = Math.random()
                let percentage = 0
                let approved = 0
                let score = 0
                if (state.submissions[state.activeCourse][index] !== undefined) {
                    let submission = state.submissions[state.activeCourse][index]
                    approved = submission.getStatus()
                    percentage = 100 - (submission.getScore() - rand * 100)
                    score = submission.getScore() - rand * 100
                }

                return (
                    <li style={{position: "relative", height:"45px"}} className={active === index && state.activeCourse === assignment.getCourseid() ? "active" : ""} key={assignment.getId()} onClick={() => {setActive(index)}}>
                        <div id="icon">
                            <i className={assignment.getIsgrouplab() ? "fa fa-users" : "fa fa-user"} title={assignment.getIsgrouplab() ? "Group Lab" : "Individual Lab"}>
                            </i>
                            
                        </div>
                        <div id="title">
                            <Link to={`/course/${state.activeCourse}/${assignment.getId()}`}>{assignment.getName()}</Link>
                        </div> 
                        {approved==1 && 
                            <i className="fa fa-check">    </i>
                        }
                        {/** The following code adds a "progress" bar below the lab in the sidebar to indicate how many % done a user is with a lab where the percentage is (100 - submission.getScore()).
                         * if the score is above the assignment score limit, the bar turns green, else it is yellow. //TODO Should look into de-spaghettifying this. */ }
                        <ProgressBar courseID={state.activeCourse} assignmentIndex={index} type="navbar" />

                        {/* <div style={{ position: "absolute", borderBottom: "1px solid green", bottom: 0, left: 0, right: `${percentage}%`, borderColor: `${ score >= assignment.getScorelimit() ? "green" : "yellow"}`, opacity:0.3 }}></div> */}
                        {/* TODO This line should perhaps be its own component */}
                    </li>
                )
            })
        return links
        }
        return []

    }
    return (
        <React.Fragment>
            {labs()}
        </React.Fragment>
    )
}

export default NavBarLabs