import React from "react"
import { Redirect } from "react-router"
import { hasEnrollment } from "../Helpers"
import { useAppState } from "../overmind"
import Alert from "./Alert"
import Courses from "./Courses"
import SubmissionsTable from "./dashboard/SubmissionsTable"


/* Dashboard for a signed in user. */
const Dashboard = (): JSX.Element => {
    const state = useAppState()

    // Users that are not enrolled in any courses are redirected to the course list.
    if (!hasEnrollment(state.enrollments)) {
        return <Redirect to={"/courses"} />
    }

    return (
        <div className='box'>
            <Alert />
            <div>
                <h1>Welcome, {state.self.Name}!</h1>
            </div>
            <SubmissionsTable />
            <Courses home />
        </div>
    )
}

export default Dashboard
