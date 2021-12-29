import React from "react"
import { useAppState } from "../../overmind"
import NavBarLink, { NavLink } from "./NavBarLink"



export const NavBarTeacher = (): JSX.Element => {
    const state = useAppState()

    const links: NavLink[] = [
        {icons: [state.pendingEnrollments.length > 0 ? {text: state.pendingEnrollments.length.toString(), classname: "badge badge-danger"} : null, {text: state.numEnrolled.toString(), classname: "badge badge-primary"}], link: {text: "Members", to: `/course/${state.activeCourse}/members`}},
        {link: {text: "Review", to: `/course/${state.activeCourse}/review`}},
        {link: {text: "Groups", to: `/course/${state.activeCourse}/groups`}},
        {link: {text: "Results", to: `/course/${state.activeCourse}/results`}},
        {link: {text: "Statistics", to: `/course/${state.activeCourse}/statistics`}},
    ]

    const teacherLinks = links.map((link, index) => { return <NavBarLink key={index} link={link.link} icons={link.icons} /> })
    return <>{teacherLinks}</>
}

export default NavBarTeacher