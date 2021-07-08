import React from "react";
import { useOvermind } from "../overmind";
import Courses from "./Courses";
import LandingPageLabTable from "./LandingPageLabTable";


/* Dashboard for a signed in user. */
const Dashboard = () => {
    const { state: {self: user} } = useOvermind()

    return(
        <div className='box'>
            <div>
                <h1>Welcome, {user.getName()}!</h1>
            </div>
            <LandingPageLabTable courseID={0}/>
            <Courses home={true} />
        </div>
    )
}

export default Dashboard