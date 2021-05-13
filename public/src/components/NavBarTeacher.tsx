import React from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"
import { Enrollment } from "../../proto/ag_pb"



export const NavBarTeacher = (props: {courseID: number}) => {

    const {state} = useOvermind()

    const pendingMembers = state.courseEnrollments[props.courseID].filter(user => user.getStatus() === Enrollment.UserStatus.PENDING).length
    const totalMembers = state.courseEnrollments[props.courseID].filter(user => user.getStatus() !== Enrollment.UserStatus.PENDING).length

    return (
        <React.Fragment>
        <li key={"members"} className="activeLabs">
            <div id="icon" className={"badge badge-danger"}>
                {pendingMembers}
            </div>
            <div id="icon" className={"badge badge-primary"}>
                {totalMembers}
            </div>
            <div id="title">
                <Link to={`/course/${state.activeCourse}/members`}>Members</Link>
            </div>
        </li>
        <li key={"review"} className="activeLabs">
            <div id="icon">

            </div>
            <div id="title">
                <Link to={`/course/${state.activeCourse}/review`}>Review</Link>
            </div>
        </li>
        <li key={"groups"} className="activeLabs">
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