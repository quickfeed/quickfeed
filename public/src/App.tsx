import React from 'react'
import { useAppState } from './overmind'
import NavBar from "./components/NavBar"
import { Switch, Route } from 'react-router-dom'
import Profile from "./components/profile/Profile"
import CoursePage from "./pages/CoursePage"
import Courses from "./components/Courses"
import AdminPage from './pages/AdminPage'
import Loading from './components/Loading'
import Dashboard from './components/Dashboard'
import AboutPage from './pages/AboutPage'
import LoginPage from './pages/LoginPage'

const App = () => {
    const state = useAppState()

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
                    <Route path="/login" component={LoginPage} />
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
                    <Route path="/" component={LoginPage} />
                </Switch>
            )
        }
    }

    return (
        <div>
            <NavBar />
            <div className="app wrapper">
                <div id={state.showFavorites ? "content" : "content-full"}>
                    {Main()}
                </div>
            </div>
        </div>
    )
}

export default App
