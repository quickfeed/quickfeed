import React from "react"
import { Route, Routes, useLocation } from "react-router"
import CourseLabs from "../components/student/CourseLabs"
import CourseLinks from "../components/CourseLinks"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alerts from "../components/alerts/Alerts"
import { useCourseID } from "../hooks/useCourseID"


const StudentPage = () => {
    const courseID = useCourseID()
    const location = useLocation()
    const root = `/course/${courseID}`

    return (
        <div className="box">
            <RedirectButton to={root} />
            <Alerts />
            <div className="row" hidden={location.pathname !== root}>
                <div className="col-md-9" >
                    <CourseLabs />
                </div>
                <CourseLinks />
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
