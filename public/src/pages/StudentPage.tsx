import React from "react"
import { Route, Routes, useLocation } from "react-router"
import CourseLabs from "../components/student/CourseLabs"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alerts from "../components/alerts/Alerts"
import { useCourseID } from "../hooks/useCourseID"
import Card from "../components/Card"
import { useAppState } from "../overmind"
import { RepositoryCards } from "../components/student/RepositoryCards"


const StudentPage = () => {
    const state = useAppState()
    const courseID = useCourseID()
    const location = useLocation()
    const root = `/course/${courseID}`
    const enrollment = state.enrollmentsByCourseID[courseID.toString()]
    const repos = state.repositories[courseID.toString()]
    const hasGroup = state.hasGroup(courseID.toString())
    const groupName = enrollment?.group ? `(${enrollment.group.name})` : ""

    const groupCard = {
        title: hasGroup ? `View Group ${groupName}` : "Create a Group",
        text: hasGroup ? "View your group." : "Create a group for this course.",
        buttonText: hasGroup ? "View Group" : "Create a Group",
        to: `${root}/group`
    }

    return (
        <div className="box">
            <RedirectButton to={root} />
            <Alerts />
            <div hidden={location.pathname !== root}>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                    <RepositoryCards repositories={repos} groupName={groupName} />
                    <Card {...groupCard} />
                </div>
                <CourseLabs />
            </div>
            <Routes>
                <Route path="/group" element={<GroupPage />} />
                <Route path="/lab/:lab" element={<Lab />} />
                <Route path="/group-lab/:lab" element={<Lab />} />
            </Routes>
        </div>
    )
}

export default StudentPage
