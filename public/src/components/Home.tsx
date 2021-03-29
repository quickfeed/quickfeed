import React, {useEffect} from "react";
import { useOvermind, useState } from "../overmind";
import { Link } from "react-router-dom";
import { getFormattedDeadline } from "../Helpers";
import LandingPageLabTable from "./LandingPageLabTable"
import { Assignment, Repository } from "../proto/ag_pb";



const Home = () => {
    const { state } = useOvermind()
    
    useEffect(() => {
    }, [])
    

    return(
        <div className='box'>
                
            {state.user.id > 0 &&
            <div>
                <h1>Welcome, {state.user.name}!</h1>
            </div>
            }
            <LandingPageLabTable courseID={0}/>           
        </div>
        )
}


export default Home;