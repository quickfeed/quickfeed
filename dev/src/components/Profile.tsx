import React, { useState } from 'react'
import { Redirect } from 'react-router'
import { useAppState } from '../overmind'
import UserProfileForm from './forms/UserProfileForm'


const Profile = () => {
    const state = useAppState()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // Flips between editable and uneditable view of user info
    const editProfile = () => {
        setEditing(!editing)
    }

    const ProfileInfo = () => {
        return (
            <div className="box">
                    <div className="card well">
                    <div className="card-header">Your Information</div>
                        <ul className="list-group list-group-flush">
                            <li className="list-group-item">Name: {state.self.getName()}</li>
                            <li className="list-group-item">Email: {state.self.getEmail()}</li>
                            <li className="list-group-item">Student ID: {state.self.getStudentid()}</li>
                        </ul>
                    </div>
                <button className="btn btn-primary" onClick={() => editProfile()}>Edit Profile</button>
            </div>
            )
    }


    if (state.self.getId() > 0) {
        return (
            <div className="box">
                <div className="jumbotron">
                    <div className="centerblock container">
                    <h1>Hi, {state.self.getName()}</h1>
                    You can edit your user information here.
                    </div>
                </div>
                <div className="container">
                {editing ? <UserProfileForm setEditing={setEditing} /> : <ProfileInfo />}
                </div>
            </div>
            )
    }
    return <Redirect to="/" />

    
}

export default Profile