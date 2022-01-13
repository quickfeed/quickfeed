import React, { useEffect } from 'react'
import { useAppState, useActions } from './overmind'
import NavBar from "./components/NavBar"
import { Switch, Route } from 'react-router-dom'
import Profile from "./components/Profile"
import CoursePage from "./pages/CoursePage"
import Courses from "./components/Courses"
import AdminPage from './pages/AdminPage'
import Loading from './components/Loading'
import Dashboard from './components/Dashboard'
import AboutPage from './pages/AboutPage'
import { isValid } from './Helpers'

const App = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        async function setup() {
            await actions.fetchUserData()
        }
        // If the user is not logged in, fetch user data to initialize the app state.
        if (!state.isLoggedIn) {
            setup()
        }
    }, [])

    // Update the state's Time object every 20 seconds.
    // After component mounts, it uses a new Time object
    useEffect(() => {
        const updateDateNow = setInterval(() => {
            actions.setTimeNow()
        }, 1200000)
        return () => clearInterval(updateDateNow)
    }, [])

    const Main = () => {
        // Determine which routes are available to the user depending on the state
        if (state.isLoading) {
            return <Loading />
        } else if (!isValid(state.self) && state.isLoggedIn) {
            // user logged in without profile information: redirect to Profile page
            return (
                <Switch>
                    <Route path="/" component={Profile} />
                    <Route path="/profile" component={Profile} />
                </Switch>
            )
        } else if (state.isLoggedIn) {
            // user logged in: show Dashboard page
            return (
                <Switch>
                    <Route path="/" exact component={Dashboard} />
                    <Route path="/about" component={AboutPage} />
                    <Route path="/profile" component={Profile} />
                    <Route path="/course/:id" component={CoursePage} />
                    <Route path="/courses" exact component={Courses} />
                    <Route path="/admin" component={AdminPage} />
                </Switch>
            )
        } else {
            //  user not logged in: show About page
            return (
                <Switch>
                    <Route path="/" component={AboutPage} />
                </Switch>
            )
        }
    }

    return (
        <div>
            <NavBar />
            <div className="app wrapper">
                <div id="content">
                    {Main()}
                </div>
            </div>
        </div>
    )

}



export default App
