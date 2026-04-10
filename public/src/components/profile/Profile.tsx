import React, { useEffect, useState } from 'react'
import { useNavigate, useLocation } from 'react-router'
import { useAppState } from '../../overmind'
import ProfileForm from './ProfileForm'
import ProfileCard from './ProfileCard'
import ProfileInfo from './ProfileInfo'
import SignupText from './SignupText'
import ProfileHero from './ProfileHero'


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
            <ProfileHero name={state.self.Name} />
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
