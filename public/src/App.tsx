import React, { Component, useEffect, useState } from "react";
import { useOvermind } from "./overmind";
import Home from './components/Home'
import Info from "./components/Info";
import NavBar from "./components/NavBar";
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom';
import Profile from "./components/Profile";
import Course from "./components/Course";
import Lab from "./components/Lab";



const App = () => {

    const [loggedIn, setLoggedIn] = useState(false)

    useEffect(() => {
        if (!loggedIn) {
            actions.getUser().then(res => {setLoggedIn(res)
                if(res){
                    actions.getEnrollmentsByUser()
                    .then(success => {
                        if (success) {
                            state.enrollments.map(enroll => {
                                actions.getAssignmentsByCourse(enroll.getCourseid()).then(success => {
                                    if (success) {
                                        actions.getSubmissions(enroll.getCourseid()).then(success => {if(success) actions.getCourses()})
                                    }
                                })
                            })
                            
                        }
                    });
                }
            }
                
            
            ) // Sets loggedIn to whatever getUser() resolves to. (fetches from /api/v1/user and resolves to true or false)
           
        }
        
        console.log("App.tsx useeffect runs")
        actions.setTheme()
        document.body.className = state.theme
    }, [])
    
    // General
    const { state, actions, effects } = useOvermind()
    return ( 
        <Router>
            <NavBar />
            <div className={state.theme+" app container"} >
                
                {!loggedIn ? ( // if not logged in, enable only the Info component to be rendered
                    <Switch>
                        <Route path="/" component={Info} />
                    </Switch>
                ) : ( // Else, enable components that require authentication
                <Switch>
                    <Route path="/" exact component={Home}/>
                    <Route path="/info" component={Info} />
                    <Route path="/profile" component={Profile} />
                    <Route path="/course/:id" component={Course} />
                </Switch>
                // Admin stuff is probably also needed here somewhere. 
                )}
                
            </div>
        </Router>
        )

}



export default App;