import React, { Component, useEffect, useState } from 'react'
import { useOvermind } from './overmind'
import Home from './components/Home'
import Info from "./components/Info";
import NavBar from "./components/NavBar";
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import Profile from "./components/Profile";
import Course from "./components/Course";
import Lab from "./components/Lab";
import Courses from "./components/Courses";



const App = () => {

    const [loggedIn, setLoggedIn] = useState(false)
    
    useEffect(() => {
        if (!loggedIn) {
             actions.setupUser().then(success => {
                if (success) {
                    setLoggedIn(true)
                }
            })
        }
        
        console.log('App.tsx useeffect runs')
        actions.setTheme()
    }, [loggedIn, setLoggedIn])

    // General
    const { state, actions } = useOvermind()
    return ( 
        <Router>
            <div className={state.theme+" app wrapper"} >
            <NavBar />
            
                <div id="content">
                {!loggedIn ? ( // if not logged in, enable only the Info component to be rendered
                    <Switch>
                        <Route path="/" component={Info} />
                    </Switch>
                ) : ( // Else if, user logged in, but has not added their information redirect to Profile
                (state.user.email.length == 0 || state.user.name.length == 0 || state.user.studentid == 0) && loggedIn ? (
                    <Switch>
                        <Route path="/" component={Profile} />
                    </Switch>
                ) : ( state.isLoading ? ( // Else render page as expected for a logged in user
                <Switch>
                    <Route path="/" exact component={Home}/>
                    <Route path="/info" component={Info} />
                    <Route path="/profile" component={Profile} />
                    <Route path="/course/:id" exact component={Course} />
                    <Route path="/courses" exact component={Courses} />
                    <Route path="/course/:id/:lab" component={Lab} />
                </Switch>
                // Admin stuff is probably also needed here somewhere. 
                ) : (
                    <h1>Loading</h1>
                )))}
                </div>
            </div>
        </Router>
        )

}



export default App