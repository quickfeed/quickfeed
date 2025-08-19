import React from 'react';
import { useAppState } from './overmind';
import NavBar from "./components/NavBar";
import { Route, Routes } from 'react-router-dom';
import Profile from "./components/profile/Profile";
import CoursePage from "./pages/CoursePage";
import Courses from "./components/Courses";
import AdminPage from './pages/AdminPage';
import Loading from './components/Loading';
import Dashboard from './components/Dashboard';
import AboutPage from './pages/AboutPage';
import LoginPage from './pages/LoginPage';
const App = () => {
    const state = useAppState();
    const Main = () => {
        if (state.isLoading) {
            return React.createElement(Loading, null);
        }
        else if (!state.isValid && state.isLoggedIn) {
            return (React.createElement(Routes, null,
                React.createElement(Route, { path: "*", element: React.createElement(Profile, null) })));
        }
        else if (state.isLoggedIn) {
            return (React.createElement(Routes, null,
                React.createElement(Route, { path: "/login", element: React.createElement(LoginPage, null) }),
                React.createElement(Route, { path: "/about", element: React.createElement(AboutPage, null) }),
                React.createElement(Route, { path: "/profile", element: React.createElement(Profile, null) }),
                React.createElement(Route, { path: "/course/:id/*", element: React.createElement(CoursePage, null) }),
                React.createElement(Route, { path: "/courses", element: React.createElement(Courses, { home: false }) }),
                React.createElement(Route, { path: "/admin/*", element: React.createElement(AdminPage, null) }),
                React.createElement(Route, { path: "*", element: React.createElement(Dashboard, null) })));
        }
        else {
            return (React.createElement(Routes, null,
                React.createElement(Route, { path: "*", element: React.createElement(LoginPage, null) })));
        }
    };
    return (React.createElement("div", null,
        React.createElement(NavBar, null),
        React.createElement("div", { className: "app wrapper" },
            React.createElement("div", { id: state.showFavorites ? "content" : "content-full" }, Main()))));
};
export default App;
