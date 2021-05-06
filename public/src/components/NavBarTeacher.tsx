import React from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"
import { Enrollment } from "../../proto/ag_pb"



export const NavBarTeacher = (props: {courseID: number}) => {

    const {state} = useOvermind()


    return (
        <React.Fragment>
        <li className="active">
            <div id="icon" className={"badge badge-danger"}>
                {state.courseEnrollments[props.courseID].filter(user => user.getStatus() === Enrollment.UserStatus.PENDING).length}
            </div>
            <div id="icon" className={"badge badge-primary"}>
            {state.courseEnrollments[props.courseID].filter(user => user.getStatus() !== Enrollment.UserStatus.PENDING).length}
            </div>
            <div id="title">
                <Link to={`/course/${state.activeCourse}/members`}>Members</Link>
            </div>
        </li>
        <li className="active">
            <div id="icon">

            </div>
            <div id="title">
                <Link to={`/course/${state.activeCourse}/review`}>Review</Link>
            </div>
        </li>
        <li className="active">
            <div id="icon">

            </div>
            <div id="title">
                <Link to={`/course/${state.activeCourse}/groups`}>Groups</Link>
            </div>
        </li>
        </React.Fragment>
    )
}

export default NavBarTeacher