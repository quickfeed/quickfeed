import React, {useCallback, useState, useEffect} from "react";
import { useOvermind } from "../overmind";

import NavBar from './NavBar'
import Info from "./Info";




const Home = () => {
    const { state, actions } = useOvermind()

    const listUsers = state.users.map(user => {
        return (
        <h3><img src={user.getAvatarurl()} width='100'></img> {user.getName()}</h3>
        )
    });

    useEffect(() => {
        actions.getUsers();
    }, [])

    if (state.user.id == -1) {
        return <Info />
    }

    return (
        <div className='box'>
            <NavBar></NavBar>
            <h1>Autograder</h1>
            {listUsers}
            {state.user.id > 0 &&
            <div>
            <h1>Welcome, {state.user.name}! Metadata: {state.users}</h1>
            <img className="avatar" src={state.user.avatarurl}></img>
            </div>
            }
            {state.user.id == -1 && <Info />}
            
        </div>

    )
}

export default Home;