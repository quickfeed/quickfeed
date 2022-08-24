import React, { useEffect } from 'react'
import { useAppState, useActions } from './overmind'
import NavBar from "./components/NavBar"
import { Switch, Route } from 'react-router-dom'
import Profile from "./components/profile/Profile"
import CoursePage from "./pages/CoursePage"
import Courses from "./components/Courses"
import AdminPage from './pages/AdminPage'
import Loading from './components/Loading'
import Dashboard from './components/Dashboard'
import AboutPage from './pages/AboutPage'
import { Color } from './Helpers'

const App = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        async function setup() {
            await actions.fetchUserData()
            // TODO: Remove this once finished testing
            actions.alert({text: "📢 ATTENTION: QuickFeed is is currently being tested for deployment in the near future.\n\nWe will wipe the database before courses are published. You will need to register again once we are done.", color: Color.RED})
        }
        // If the user is not logged in, fetch user data to initialize the app state.
        if (!state.isLoggedIn) {
            setup()
        }
    }, [])

    const Main = () => {
        // Determine which routes are available to the user depending on the state
        if (state.isLoading) {
            return <Loading />
        } else if (!state.isValid && state.isLoggedIn) {
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
