import React, {useCallback, useState, useEffect} from "react";
import { useOvermind } from "../overmind";

import NavBar from './NavBar'
import Info from "./Info";
import { Enrollment } from "../proto/ag_pb";
import { Link } from "react-router-dom";




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

    const months = ["January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December"];
    const listAssignments = state.assignments.map(assignment => {
        const deadline = new Date(assignment.getDeadline())
        return (
            <h2>{assignment.getName()} Deadline: {deadline.getDate()} {months[deadline.getMonth()]} {deadline.getFullYear()} by {deadline.getHours()}:{deadline.getMinutes()} </h2>
        )
    })

    useEffect(() => {
        actions.getEnrollmentsByUser()
        .then(success => {
            if (success) {
                actions.getAssignments()
            }
        });

        console.log(state.enrollments)
    }, [])

    return (
        <div className='box'>
            <h1>Autograder</h1>
                
            {state.user.id > 0 &&
            <div>
            <h1>Welcome, {state.user.name}! Current theme: {state.theme}</h1>
            <img className="avatar" src={state.user.avatarurl}></img>
            </div>
            }
            <a><button>Courses</button></a>
            {listCourses}
            {listAssignments}
        </div>
        )
}


export default Home;