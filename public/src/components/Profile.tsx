import { useState } from "react"
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
            <div className="box">
                <h1>Hello {state.user.name}!</h1>
                <h2>Your email is {state.user.email}</h2>
                <h3>Your student ID is {state.user.studentid}</h3>
                <h4>You are {state.user.isadmin ? 'an' : 'not an' } admin</h4>
                <h5><img src={state.user.avatarurl} width="20%"></img></h5>
                <button onClick={() => editProfile()}>Edit Profile</button>
            </div>
            )
        } 
        // Render editable user information
        else {
            
            return ( 
                <div className="box">
                <form onSubmit={e => {e.preventDefault(); submitHandler()}}>
                    <h1><input name="name" type="text" value={user.name} onChange={handleChange} /></h1>
                    <h2><input name="email" type="text" value={user.email} onChange={handleChange} /></h2>
                    <h3><input name="studentid" type="text" value={user.studentid} onChange={handleChange} /></h3>
                    <input type="submit" value="Submit" />
                </form>
                </div>
            )
        }
    }
    // If the user does not have a valid ID (-1)
    // TODO: Redirect ? Not needed currently as all components are disabled except Info if not logged in. Could change
    return <h1>Not logged in.</h1>

    
}

export default Profile