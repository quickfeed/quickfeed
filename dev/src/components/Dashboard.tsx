import React from "react";
import { Redirect } from "react-router";
import { hasEnrollment } from "../Helpers";
import { useAppState } from "../overmind";
import Courses from "./Courses";
import LandingPageLabTable from "./LandingPageLabTable";


/* Dashboard for a signed in user. */
const Dashboard = (): JSX.Element => {
    const state = useAppState()

    // New users logging in are redirected to courses to ease enrollment
    if (!hasEnrollment(state.enrollments)) {
        return <Redirect to={"/courses"}></Redirect>
    }

    return(
        <div className='box'>
            <div>
                <h1>Welcome, {state.self.getName()}!</h1>
            </div>
            <LandingPageLabTable courseID={0}/>
            <Courses home={true} />
        </div>
    )
}

export default Dashboard