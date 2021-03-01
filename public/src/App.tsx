import React, { Component, useEffect, useState } from "react";
import { useOvermind } from "./overmind";
import Home from './components/Home'
import './App.css'
import Info from "./components/Info";
import NavBar from "./components/NavBar";
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import Profile from "./components/Profile";
import Course from "./components/Course";



const App = () => {

    const [loggedIn, setLoggedIn] = useState(false)

    useEffect(() => {
        actions.getUser()
        .then(res => setLoggedIn(res)) // Sets loggedIn to whatever getUser() resolves to. (fetches from /api/v1/user and resolves to true or false)
        actions.setTheme()
        document.body.className = state.theme
    }, [loggedIn, setLoggedIn])

    // General
    const { state, actions, effects } = useOvermind()
    return ( 
        <Router>
            <div className={state.theme+" app"} >
                <NavBar />
                {!loggedIn ? ( // if not logged in, enable only the Info component to be rendered
                    <Switch>
                        <Route path="/" component={Info} />
                    </Switch>
                ) : ( // Else, enable components that require authentication
                <Switch>
                    <Route path="/" exact component={Home} on/>
                    <Route path="/info" component={Info} />
                    <Route path="/profile" component={Profile} />
                    <Route path="/course/:id" component={Course} />
                </Switch>
                )}
                
            </div>
        </Router>
        )

}



export default App;