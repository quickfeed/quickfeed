import React from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"
import { Enrollment } from "../proto/ag_pb"



export const NavBarTeacher = (props: {courseID: number}) => {

    const {state} = useOvermind()


    return (
        <li className="active">
            <div id="icon" className={"badge badge-danger"}>
                {state.courseEnrollments[props.courseID].filter(user => user.getStatus() === Enrollment.UserStatus.PENDING).length}
            </div>
            <div id="title">
                <Link to={`/course/${state.activeCourse}/members`}>Members</Link>
            </div>
        </li>
    )
}

export default NavBarTeacher