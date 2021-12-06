import React, { useState } from 'react'
import { Redirect } from 'react-router'
import { useAppState } from '../overmind'
import UserProfileForm from './forms/UserProfileForm'
import ProfileInfo from './ProfileInfo'


const Profile = (): JSX.Element => {
    const state = useAppState()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // User is not logged in if self.getId() <= 0
    if (state.self.getId() > 0) {
        return (
            <div>
                <div className="jumbotron">
                    <div className="centerblock container">
                    <h1>Hi, {state.self.getName()}</h1>
                    You can edit your user information here.
                    </div>
                </div>
                <div className="container">
                {editing ? <UserProfileForm setEditing={setEditing} /> : <ProfileInfo setEditing={setEditing} />}
                </div>
            </div>
            )
    }
    return <Redirect to="/" />

    
}

export default Profile