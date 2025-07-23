import React from 'react'
import { useAppState } from './overmind'
import NavBar from "./components/NavBar"
import { Route, Routes } from 'react-router-dom'
import Profile from "./components/profile/Profile"
import CoursePage from "./pages/CoursePage"
import Courses from "./components/Courses"
import AdminPage from './pages/AdminPage'
import Loading from './components/Loading'
import Dashboard from './components/Dashboard'
import AboutPage from './pages/AboutPage'
import Settings from './components/settings/Settings'
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
                <Routes>
                    <Route path="*" element={<Profile />} />
                </Routes>
            )
        } else if (state.isLoggedIn) {
            // user logged in: show Dashboard page
            return (
                <Routes>
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/about" element={<AboutPage />} />
                    <Route path="/profile" element={<Profile />} />
                    <Route path="/settings" element={<Settings />} />
                    <Route path="/course/:id/*" element={<CoursePage />} />
                    <Route path="/courses" element={<Courses home={false} />} />
                    <Route path="/admin/*" element={<AdminPage />} />
                    <Route path="*" element={<Dashboard />} />
                </Routes>
            )
        } else {
            //  user not logged in: show About page
            return (
                <Routes>
                    <Route path="*" element={<LoginPage />} />
                </Routes>
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
