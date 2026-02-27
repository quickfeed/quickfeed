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
        <div className="min-h-screen">
            <div className="hero">
                <div className="hero-content text-center">
                    <div className="max-w-2xl">
                        <h1 className="text-5xl font-bold mb-4">Hi, {state.self.Name}</h1>
                        <p className="text-lg mb-2">You can edit your user information here.</p>
                        <div className="alert alert-warning mt-4">
                            <i className="fa fa-exclamation-triangle"></i>
                            <span><strong>Use your real name as it appears on Canvas</strong> to ensure that approvals are correctly attributed.</span>
                        </div>
                    </div>
                </div>
            </div>
            <div className="container mx-auto px-4 py-12 flex justify-center">
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
