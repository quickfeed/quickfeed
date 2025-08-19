import React, { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router';
import { useAppState } from '../../overmind';
import ProfileForm from './ProfileForm';
import ProfileCard from './ProfileCard';
import ProfileInfo from './ProfileInfo';
import SignupText from './SignupText';
const Profile = () => {
    const state = useAppState();
    const navigate = useNavigate();
    const location = useLocation();
    const [editing, setEditing] = useState(false);
    useEffect(() => {
        if (!state.isLoggedIn) {
            navigate("/");
        }
        else if (!state.isValid && location.pathname === "/") {
            navigate("/profile");
        }
    }, [state.isLoggedIn, state.isValid, location.pathname, navigate]);
    useEffect(() => {
        if (!state.isValid) {
            setEditing(true);
        }
    }, [state.isValid]);
    return (React.createElement("div", null,
        React.createElement("div", { className: "banner jumbotron" },
            React.createElement("div", { className: "centerblock container" },
                React.createElement("h1", null,
                    "Hi, ",
                    state.self.Name),
                React.createElement("p", null, "You can edit your user information here."),
                React.createElement("p", null,
                    React.createElement("span", { className: 'font-weight-bold' }, "Use your real name as it appears on Canvas"),
                    " to ensure that approvals are correctly attributed."))),
        React.createElement("div", { className: "container" },
            React.createElement(ProfileCard, null, editing ?
                React.createElement(ProfileForm, { setEditing: setEditing }, state.isValid ? null : React.createElement(SignupText, null))
                : React.createElement(ProfileInfo, { setEditing: setEditing })))));
};
export default Profile;
