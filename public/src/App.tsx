import React, { useEffect, useState } from 'react'
import { useOvermind } from './overmind'
import Home from './components/Home'
import About from "./components/About";
import NavBar from "./components/NavBar";
import { Switch, Route } from 'react-router-dom';
import Profile from "./components/Profile";
import CoursePage from "./components/CoursePage"
import Courses from "./components/Courses";
import Alert from './components/Alert';
import Admin from './components/Admin';
import Loading from './components/Loading';





const App = () => {
    const { state, actions } = useOvermind()
    const [loggedIn, setLoggedIn] = useState(false)
    useEffect( () => {
        async function setup() {
            const result = await actions.fetchUserData()
            setLoggedIn(result)
        }
        if (!loggedIn) {
            setup()
        }
    }, [loggedIn,setLoggedIn])

    // This is just to Update the Time object in state, every 20 minutes (after mount, it mounts with a new dateobject)
    useEffect(()=> {
        let updateDateNow = setInterval(()=>{
            actions.setTimeNow()
        },1200000)
        return() => clearInterval(updateDateNow)
    },[])


    // General
    
    return  (
        <div> 
            <NavBar />
                <div className={state.theme+" app wrapper"} >
                    <div id="content">
                    <Alert />
                        {state.isLoading ? ( // If state.isLoading
                            <Loading />
                            ) : ( // Else if, user logged in, but has not added their information redirect to Profile
                                (state.user.email.length == 0 || state.user.name.length == 0 || state.user.studentid == 0) && loggedIn ? (
                                    <Switch>
                                        <Route path="/" component={Profile} />
                                    </Switch>
                            ) : (loggedIn ? ( // Else render page as expected for a logged in user
                                <Switch>
                                    <Route path="/" exact component={Home} />
                                    <Route path="/about" component={About} />
                                    <Route path="/profile" component={Profile} />
                                    <Route path="/course/:id" component={CoursePage} />
                                    <Route path="/courses" exact component={Courses} />
                                    <Route path="/admin" component={Admin} />
                                </Switch>
                                // Admin stuff is probably also needed here somewhere. 
                            ) : (
                                <Switch>
                                    <Route path="/" component={About} />
                                </Switch>
                        )))}
                </div>
            </div>
        </div>
    )

}



export default App