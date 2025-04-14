import React from "react"
import { isManuallyGraded } from "../../Helpers"
import { useAppState } from "../../overmind"
import NavBarLink, { NavLink } from "./NavBarLink"


const NavBarTeacher = () => {
    const state = useAppState()
    const pending = state.pendingEnrollments.length > 0 ? { text: state.pendingEnrollments.length.toString(), classname: "badge badge-danger" } : null
    const enrolled = { text: state.numEnrolled.toString(), classname: "badge badge-primary" }
    const pendingGroups = state.pendingGroups.length > 0 ? { text: state.pendingGroups.length.toString(), classname: "badge badge-danger" } : null
    const approvedGroups = { text: state.numGroups.toString(), classname: "badge badge-primary" }
    const courseHasManualGrading = state.assignments[state.activeCourse.toString()]?.some(assignment => isManuallyGraded(assignment.reviewers))

    const links: NavLink[] = [
        { link: { text: "Results", to: `/course/${state.activeCourse}/results` } },
        { link: { text: "Members", to: `/course/${state.activeCourse}/members` }, icons: [pending, enrolled] },
        { link: { text: "Groups", to: `/course/${state.activeCourse}/groups` }, icons: [pendingGroups, approvedGroups] },
    ]

    if (courseHasManualGrading) {
        links.unshift({ link: { text: "Review", to: `/course/${state.activeCourse}/review` } })
    }

    const teacherLinks = links.map((link) => { return <NavBarLink key={link.link.text} link={link.link} icons={link.icons} /> })
    return <>{teacherLinks}</>
}

export default NavBarTeacher
