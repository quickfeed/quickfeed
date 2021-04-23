import React, { useState } from 'react'
import { useOvermind } from '../overmind'
import UserProfileForm from './forms/UserProfileForm'

const Profile = () => {
    const { state } = useOvermind()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // Flips between editable and uneditable view of user info
    const editProfile = () => {
        setEditing(!editing)
    }

    // Returns if the user has a valid ID
    if (state.user.id > 0) {
        // Render user information
        if(editing === false) {
        return (
            <div className="box" style={{color: "black"}}>
                <div className="jumbotron"><div className="centerblock container"><h1>Hi, {state.user.name}</h1>You can edit your user information here.</div></div>
                
                    <div className="card well" style={{width: "400px"}}>
                    <div className="card-header">Your Information</div>
                        <ul className="list-group list-group-flush">
                            <li className="list-group-item">Name: {state.user.name}</li>
                            <li className="list-group-item">Email: {state.user.email}</li>
                            <li className="list-group-item">Student ID: {state.user.studentid}</li>
                        </ul>
                    </div>
                <button className="btn btn-primary" onClick={() => editProfile()}>Edit Profile</button>
            </div>
            )
        } 
        // Render editable user information.
        else {
            return (
                <UserProfileForm editing={editing} setEditing={setEditing} />
            )
        }
    }
    return <h1>Not logged in.</h1>

    
}

export default Profile