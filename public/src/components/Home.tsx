import React from "react";
import { useOvermind } from "../overmind";
import SubmissionsTable from "./SubmissionsTable"


/* Dashboard for a signed in user. */
const Home = () => {
    const { state: {user} } = useOvermind()

    return(
        <div className='box'>       
            <div>
                <h1>Welcome, {user.name}!</h1>
            </div>
            <SubmissionsTable courseID={0}/>           
        </div>
    )
}

export default Home