import { useState } from "react"
import { useOvermind } from "../overmind"
import { state } from "../overmind/state"
import { User } from "../proto/ag_pb"




const Profile = () => {
    const { state } = useOvermind()
    // Holds a local state to check whether the user is editing their user information or not
    const [editing, setEditing] = useState(false)

    // experimenting with object state
    const [user, setUser] = useState(User)

    // Flips the above local state.
    const editProfile = () => {
        setEditing(!editing)

        setUser(user)
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
            // TODO: Make a form and on save, push it into a gRPC call where the user is changed server-side. (Change state, then API call to server with state.user as input?)
            return ( 
                <form>
                    <label>
                    Name:
                    <input type="text" value={state.user.name} />
                    </label>
                    <input type="submit" value="Submit" />
                    <button onClick={() => editProfile()}>Editing profile</button> 
                </form>
                
            )
        }
    }
    // If the user does not have a valid ID (-1)
    // TODO: Redirect ? Not needed currently as all components are disabled except Info if not logged in. Could change
    return <h1>Not logged in.</h1>

    
}

export default Profile