import React from "react"
import { Redirect } from "react-router"
import { hasEnrollment } from "../Helpers"
import { useAppState } from "../overmind"
import Alerts from "./alerts/Alerts"
import Courses from "./Courses"
import SubmissionsTable from "./dashboard/SubmissionsTable"


/* Dashboard for a signed in user. */
const Dashboard = () => {
    const state = useAppState()

    // Users that are not enrolled in any courses are redirected to the course list.
    if (!hasEnrollment(state.enrollments)) {
        return <Redirect to={"/courses"} />
    }

    return (
        <div className="mt-5">
            <Alerts />
            <Courses home />
            <SubmissionsTable />
        </div>
    )
}

export default Dashboard
