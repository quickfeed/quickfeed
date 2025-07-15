import React, { useEffect, useState } from 'react'
import { useNavigate, useLocation } from 'react-router'
import { useAppState } from '../../overmind'
import ProfileForm from './ProfileForm'
import ProfileCard from './ProfileCard'
import ProfileInfo from './ProfileInfo'
import SignupText from './SignupText'


const Profile = () => {
    const state = useAppState()
    const navigate = useNavigate()
    const location = useLocation()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // Redirect from "/" to "/profile" when user object is invalid
    useEffect(() => {
        if (!state.isLoggedIn) {
            navigate("/")
        } else if (!state.isValid && location.pathname === "/") {
            navigate("/profile")
        }
    }, [state.isLoggedIn, state.isValid, location.pathname, navigate])

    // Default to edit mode if user object does not contain valid information
    useEffect(() => {
        if (!state.isValid) {
            setEditing(true)
        }
    }, [state.isValid])

    return (
        <div>
            <div className="banner jumbotron">
                <div className="centerblock container">
                    <h1>Hi, {state.self.Name}</h1>
                    <p>You can edit your user information here.</p>
                    <p><span className='font-weight-bold'>Use your real name as it appears on Canvas</span> to ensure that approvals are correctly attributed.</p>
                </div>
            </div>
            <div className="container">
                <ProfileCard>
                    {editing ?
                        <ProfileForm setEditing={setEditing} >
                            {state.isValid ? null : <SignupText />}
                        </ProfileForm>
                        : <ProfileInfo setEditing={setEditing} />
                    }
                </ProfileCard>
            </div>
        </div>
    )
}

export default Profile
