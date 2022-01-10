import React, { useEffect, useState } from 'react'
import { Redirect, useHistory } from 'react-router'
import { useAppState } from '../overmind'
import ProfileForm from './forms/ProfileForm'
import ProfileCard from './profile/ProfileCard'
import ProfileInfo from './profile/ProfileInfo'
import SignupText from './profile/SignupText'


const Profile = (): JSX.Element => {
    const state = useAppState()
    const history = useHistory()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // Redirect from "/" to "/profile" when user object is invalid
    if (!state.isValid && history.location.pathname == "/") {
        history.push("/profile")
    }

    // Default to edit mode if user object does not contain valid information
    useEffect(() => {
        if (!state.isValid) {
            setEditing(true)
        }
    })

    if (state.isLoggedIn) {
        return (
            <div>
                <div className="jumbotron">
                    <div className="centerblock container">
                        <h1>Hi, {state.self.getName()}</h1>
                        You can edit your user information here.
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
    return <Redirect to="/" />
}

export default Profile
