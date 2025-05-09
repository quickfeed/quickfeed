import React from "react"
import { Route, Routes } from "react-router"
import { getCourseID } from "../Helpers"
import CourseLabs from "../components/student/CourseLabs"
import CourseLinks from "../components/CourseLinks"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alerts from "../components/alerts/Alerts"
import { useLocation } from "react-router"


const StudentPage = () => {
    const courseID = getCourseID()
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
