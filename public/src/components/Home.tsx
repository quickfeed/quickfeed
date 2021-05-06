import React, {useEffect} from "react";
import { useOvermind, useState } from "../overmind";
import { Link } from "react-router-dom";
import { getFormattedDeadline } from "../Helpers";
import LandingPageLabTable from "./LandingPageLabTable"
import { Assignment, Repository } from "../../proto/ag_pb";



const Home = () => {
    const { state } = useOvermind()
    
    useEffect(() => {
    }, [])
    

    return(
        <div className='container-fluid box'>
            <div className="row">
                {state.user.id > 0 &&
                    <h1>Welcome, {state.user.name}!</h1>
                }
                <div className="col-xl-10">
                    <LandingPageLabTable courseID={0}/>   
                </div>
                    
            </div>
             
        </div>
    )
}


export default Home