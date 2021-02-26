import React, {useCallback, useState, useEffect} from "react";
import { useOvermind } from "../overmind";

import NavBar from './NavBar'
import Info from "./Info";
import { Enrollment } from "../proto/ag_pb";




const Home = () => {
    const { state, actions } = useOvermind()

    const listUsers = state.users.map(user => {
        return (
        <h3><img src={user.getAvatarurl()} width='100'></img> {user.getName()}</h3>
        )
    });

    const listCourses = state.courses.map(course => {
        return (
            <h5>{course.getName()}</h5>
        )
    })

    const handleClick = {

    }

    useEffect(() => {
        actions.getUsers();
        actions.getCourses();
        
    }, [])

    if (state.currentPage === "info") {
        return <Info />
    }

    if (state.currentPage === "home") {
        return (
            <div className='box'>
                <h1>Autograder</h1>
                
                {state.user.id > 0 &&
                <div>
                <h1>Welcome, {state.user.name}! Metadata: {state.users}</h1>
                <img className="avatar" src={state.user.avatarurl}></img>
                </div>
                }
                {state.user.id == -1 && <Info />}
                <a><button onClick={() => actions.setCurrentPage("info")}>Courses</button></a>
                {listCourses}
            </div>
            )
    }
    return (<h1>404</h1>)
}

export default Home;