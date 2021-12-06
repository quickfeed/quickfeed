import React, { useEffect, useState } from 'react'
import { useAppState, useActions } from './overmind'
import NavBar from "./components/NavBar";
import { Switch, Route } from 'react-router-dom';
import Profile from "./components/Profile";
import CoursePage from "./pages/CoursePage"
import Courses from "./components/Courses";
import Alert from './components/Alert';
import AdminPage from './pages/AdminPage';
import Loading from './components/Loading';
import { isValid } from './Helpers';
import Dashboard from './components/Dashboard';
import AboutPage from './pages/AboutPage';





const App = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
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
        const updateDateNow = setInterval(()=>{
            actions.setTimeNow()
        },1200000)
        return() => clearInterval(updateDateNow)
    },[])

    return  (
        <div> 
            <NavBar />
            <div className={state.theme+" app wrapper"} >
                <div id="content">
                <Alert />
                    {state.isLoading ? ( // If state.isLoading
                        <Loading />
                        ) : ( // Else if, user logged in, but has not added their information redirect to Profile
                            !isValid(state.self) && loggedIn ? (
                                <Switch>
                                    <Route path="/" component={Profile} />
                                </Switch>
                        ) : (loggedIn ? ( // Else render page as expected for a logged in user
                            <Switch>
                                <Route path="/" exact component={Dashboard} />
                                <Route path="/about" component={AboutPage} />
                                <Route path="/profile" component={Profile} />
                                <Route path="/course/:id" component={CoursePage} />
                                <Route path="/courses" exact component={Courses} />
                                <Route path="/admin" component={AdminPage} />
                            </Switch>
                        ) : (
                            <Switch>
                                <Route path="/" component={AboutPage} />
                            </Switch>
                    )))}
                </div>
            </div>
        </div>
    )

}



export default App