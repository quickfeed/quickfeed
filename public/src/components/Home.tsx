import React, {useEffect} from "react";
import { useOvermind } from "../overmind";
import { Link } from "react-router-dom";
import { getFormattedDeadline } from "../Helpers";
import LandingPageLabTable from "./LandingPageLabTable"
import { Assignment } from "../proto/ag_pb";



const Home = () => {
    const { state, actions } = useOvermind()

    const listUsers = state.users.map(user => {
        return (
        <h3><img src={user.getAvatarurl()} width='100'></img> {user.getName()}</h3>
        )
    });

    const listCourses = state.enrollments.map(enrollment => {
        return (
            <h5 key={enrollment.getCourseid()}>
                <Link to={`course/${enrollment.getCourseid()}`}>{enrollment.getCourse()?.getName()}</Link>
            </h5>
        )
    })
    
    const listAssignments = Object.values(state.assignments).map(assignmentArray =>{
        const abc = assignmentArray.map(assignment => {
            return(
                <h2>{assignment.getName()} Deadline: {getFormattedDeadline(assignment.getDeadline())} </h2>
            )
        })
        return abc
    })
    
    
    
    

    useEffect(() => {
        actions.getEnrollmentsByUser()
        .then(success => {
            if (success) {
                actions.getAssignments()
                state.enrollments.map(enrol =>{
                    actions.getSubmissions(enrol.getCourseid())
                })
            }
        });
    }, [])


    return (
        <div className='box'>
            <h1>Autograder</h1>
                
            {state.user.id > 0 &&
            <div>
            <h1>Welcome, {state.user.name}!</h1>
            <img className="avatar" src={state.user.avatarurl}></img>
            </div>
            }
            <a><button>Courses</button></a>
            {listCourses}
            {listAssignments}
            <LandingPageLabTable />
        </div>
        )
}


export default Home;