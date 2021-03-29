
import React, { useState } from "react"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"
import { User } from "../proto/ag_pb"




const Profile = () => {
    const { state, actions } = useOvermind()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // Local state holding information to be changed by user
    const [user, setUser] = useState({'name': state.user.name, 'email': state.user.email, "studentid": state.user.studentid})
    
    // Flips between editable and uneditable view of user info
    const editProfile = () => {
        setEditing(!editing)
    }

    // Updates local user state on change in an input field
    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const { name, value } = event.currentTarget
        setUser(prevState => ({
            ...prevState,
            [name]: value
        }))
    }

    // Sends off the edited (or not) information to the server. ((Could change actions.changeUser to take (username, email, studentid) as args and create the new user object in the action, not sure what's best))
    const submitHandler = () => {
        const changedUser = new User()
        changedUser.setId(state.user.id)
        changedUser.setName(user.name)
        changedUser.setEmail(user.email)
        changedUser.setStudentid(user.studentid.toString())
        actions.changeUser(changedUser)
        // Flip back to the uneditable view
        setEditing(false)
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
                <div className="box">
                    <div className="jumbotron"><div className="centerblock container"><h1>Hi, {state.user.name}</h1>You can edit your user information here.</div></div>
                    <form className="form-group well" style={{width: "400px"}} onSubmit={e => {e.preventDefault(); submitHandler()}}>
                        <label htmlFor={"name"}>Name</label>
                        <input className="form-control" name="name" type="text" value={user.name} onChange={handleChange} />
                        <label htmlFor={"email"}>Email</label>
                        <input className="form-control" name="email" type="text" value={user.email} onChange={handleChange} />
                        <label htmlFor={"studentid"}>Student ID</label>
                        <input className="form-control" name="studentid" type="text" value={user.studentid} onChange={handleChange} />
                        <input className="btn btn-primary" type="submit" value="Save" style={{marginTop:"20px"}}/>
                </form>
                </div>
            )
        }
    }
    return <h1>Not logged in.</h1>

    
}

export default Profile