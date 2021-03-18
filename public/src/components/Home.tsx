import React, {useEffect} from "react";
import { useOvermind, useState } from "../overmind";
import { Link } from "react-router-dom";
import { getFormattedDeadline } from "../Helpers";
import LandingPageLabTable from "./LandingPageLabTable"
import { Assignment } from "../proto/ag_pb";



const Home = () => {
    const { state, actions } = useOvermind()
    

    const listCourses = state.enrollments.map(enrollment => {
        return (
            <h5 key={enrollment.getCourseid()}>
                <Link to={`course/${enrollment.getCourseid()}`}>{enrollment.getCourse()?.getName()}</Link>
            </h5>
        )
    })
    
    
    
    useEffect(() => {
        actions.getEnrollmentsByUser()
        .then(success => {
            if (success) {
                state.enrollments.map(enrol =>{
                    actions.getAssignmentsByCourse(enrol.getCourseid()).then(success =>{
                        if (success) {
                            actions.getSubmissions(enrol.getCourseid())
                        }
                    })
                    
                })
            }
        });
    }, [])
    
    if(true){
        return(
            <div className='box'>
            <h1>Autograder</h1>
                
            {state.user.id > 0 &&
            <div>
            <h1>Welcome, {state.user.name}!</h1>
            <img className="avatar img-thumbnail" src={state.user.avatarurl}></img>
            </div>
            }
            <a><button>Courses</button></a>
            {listCourses}
            {Object.keys(state.assignments).length>0 &&
            <a>test for state</a>
            }
            <LandingPageLabTable />           
        </div>
        )
    }
    
    return (
        <div className='box'>
            <h1>loading</h1>
        </div>
        )
}


export default Home;