import React from "react"
import { useAppState } from "../../overmind"
import { Enrollment } from "../../../proto/ag/ag_pb"
import NavBarLink, { NavLink } from "./NavBarLink"



export const NavBarTeacher = (props: {courseID: number}): JSX.Element => {

    const state = useAppState()

    const pendingMembers = state.courseEnrollments[props.courseID].filter(user => user.getStatus() === Enrollment.UserStatus.PENDING).length
    const totalMembers = state.courseEnrollments[props.courseID].filter(user => user.getStatus() !== Enrollment.UserStatus.PENDING).length

    const links: NavLink[] = [
        {icons: [pendingMembers > 0 ? {text: pendingMembers.toString(), classname: "badge badge-danger"} : null, {text: totalMembers.toString(), classname: "badge badge-primary"}], link: {text: "Members", to: `/course/${state.activeCourse}/members`}},
        {link: {text: "Review", to: `/course/${state.activeCourse}/review`}},
        {link: {text: "Groups", to: `/course/${state.activeCourse}/groups`}},
        {link: {text: "Results", to: `/course/${state.activeCourse}/results`}},
        {link: {text: "Statistics", to: `/course/${state.activeCourse}/statistics`}},
    
    ]

    return (
        <React.Fragment>
            {links.map((link, index) => { return <NavBarLink key={index} link={link.link} icons={link.icons} /> })}
        </React.Fragment>
    )
}

export default NavBarTeacher