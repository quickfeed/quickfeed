import React, { Component, useEffect, useState } from 'react'
import { useOvermind } from './overmind'
import Home from './components/Home'
import Info from './components/Info'
import NavBar from './components/NavBar'
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom'
import Profile from './components/Profile'
import Course from './components/Course'
import Lab from './components/Lab'



const App = () => {

    const [loggedIn, setLoggedIn] = useState(false)
    
    useEffect(() => {
        if (!loggedIn) {
            actions.setupUser()
            .then(res => setLoggedIn(res)) // Sets loggedIn to whatever getUser() resolves to. (fetches from /api/v1/user and resolves to true or false)
        }
        
        console.log('App.tsx useeffect runs')
        actions.setTheme()
        document.body.className = state.theme
    }, [])
    
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
                state.user.email.length == 0 || state.user.name.length == 0 || state.user.studentid == 0 ? (
                    <Switch>
                        <Route path="/" component={Profile} />
                    </Switch>
                ) : ( // Else render page as expected for a logged in user
                <Switch>
                    <Route path="/" exact component={Home}/>
                    <Route path="/info" component={Info} />
                    <Route path="/profile" component={Profile} />
                    <Route path="/course/:id" component={Course} />
                </Switch>
                // Admin stuff is probably also needed here somewhere. 
                ))}
                </div>
            </div>
        </Router>
        )

}



export default App