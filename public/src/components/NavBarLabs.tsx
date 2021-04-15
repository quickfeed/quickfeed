import React, { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"


const NavBarLabs = () => {
    const {state} = useOvermind()

    const [active, setActive] = useState(-1)
    
    useEffect(() => {
    }, [state.activeCourse])

    const labs = (): JSX.Element[] => { 
        
        if(state.submissions !== undefined && state.activeCourse > 0) {

            let links = state.submissions[state.activeCourse]?.filter(submission => submission.getAssignmentid() !== 0).map((submission, index) => {
                let assignment = state.assignments[state.activeCourse][index]
                let rand = Math.random()
                let percentage = 100 - (submission.getScore() - rand * 100)
                return (
                    <li style={{position: "relative"}} className={active === index && state.activeCourse === assignment.getCourseid() ? "active" : ""} key={assignment.getId()} onClick={() => {setActive(index)}}>
                        <div id="icon">
                            <i className={assignment.getIsgrouplab() ? "fa fa-users" : "fa fa-user"} title={assignment.getIsgrouplab() ? "Group Lab" : "Individual Lab"}>
                            </i>
                        </div>
                        <div id="title">
                            <Link to={`/course/${state.activeCourse}/${submission.getAssignmentid()}`}>{assignment.getName()}</Link>
                        </div> 
                        
                        {/** The following code adds a "progress" bar below the lab in the sidebar to indicate how many % done a user is with a lab where the percentage is (100 - submission.getScore()).
                         * if the score is above the assignment score limit, the bar turns green, else it is yellow. */ }
                        <div style={{position: "absolute", borderBottom: "1px solid green", bottom: 0, left: 0, right: `${percentage}%`, borderColor: `${ submission.getScore() - rand * 100 >= assignment.getScorelimit() ? "green" : "yellow"}`, opacity:0.3 }}></div>
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