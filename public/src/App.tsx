import React, { Component, useEffect, useState } from "react";
import { useOvermind } from "./overmind";
import Todos from './components/Todos'
import './App.css'
import TodoCounter from "./components/TodoCounter";

const App = () => {

    const [loggedIn, setLoggedIn] = useState(false)

    useEffect(() => {
        actions.getUser()
        .then(res => setLoggedIn(true)) // .then(setLoggedIn(true))

        // TODO Change action.getUser to a promise, add conditional rendering that only renders after getUser() has finished.
        // Change local state const [loggedIn, setLoggedIn] = useState(); to loggedIn when getUser() finished
        // Spinny loading thingy until done?
        // {loggedIn == false && <Loading></Loading}
        // {loggedIn == true && <UserPage></UserPage>}

    }, [loggedIn, setLoggedIn])

    // General
    const { state, actions, effects, reaction } = useOvermind()
    if (loggedIn == true) {
        return <Todos />
    }
    return <TodoCounter />
}



export default App;