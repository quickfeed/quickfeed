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
import LoginPage from './pages/LoginPage'
import CourseCodeRedirect from "./components/CourseCodeRedirect"
import Alerts from './components/alerts/Alerts'

const App = () => {
    const state = useAppState()

    let routes: React.ReactNode
    if (!state.isValid && state.isLoggedIn) {
        routes = (
            <Routes>
                <Route path="*" element={<Profile />} />
            </Routes>
        )
    } else if (state.isLoggedIn) {
        routes = (
            <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/about" element={<AboutPage />} />
                <Route path="/profile" element={<Profile />} />
                <Route path="/course/:id/*" element={<CoursePage />} />
                <Route path="/courses" element={<Courses home={false} />} />
                <Route path="/admin/*" element={<AdminPage />} />
                { /* Redirect course codes to the course page, if no course found fall through to next route */}
                <Route path="/:code" element={<CourseCodeRedirect />} />
                <Route path="*" element={<Dashboard />} />
            </Routes>
        )
    } else {
        routes = (
            <Routes>
                <Route path="*" element={<LoginPage />} />
            </Routes>
        )
    }

    return (
        <div>
            <NavBar />
            <Alerts />
            {state.isLoading && <Loading />}
            <div className="app wrapper" hidden={state.isLoading}>
                <div className={`
                        transition-[margin] duration-200 ease-in-out
                        mt-8 mr-8 w-full
                        ${state.showFavorites ? "ml-64 md:ml-72" : "ml-8"}
                    `}>
                    {routes}
                </div>
            </div>
        </div>
    )
}

export default App
