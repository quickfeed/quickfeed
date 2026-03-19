import React from "react"
import { Navigate } from "react-router"
import { hasEnrollment } from "../Helpers"
import { useAppState } from "../overmind"
import Courses from "./Courses"

/* Dashboard for a signed in user. */
const Dashboard = () => {
    const state = useAppState()

    // Users that are not enrolled in any courses are redirected to the course list.
    if (!state.isLoading && !hasEnrollment(state.enrollments)) {
        return <Navigate to={"/courses"} />
    }

    return (
        <div className="mt-5">
            <Courses home />
        </div>
    )
}

export default Dashboard
