import { useOvermind } from "../overmind"
import { state } from "../overmind/state"




const Profile = () => {
    const { state } = useOvermind()
    return (
        <div className="box">
            <h1>Hello {state.user.name}!</h1>
            <h2>Your email is {state.user.email}</h2>
            <h3>Your student ID is {state.user.studentid}</h3>
            <h4>You are {state.user.isadmin ? 'an' : 'not an' } admin</h4>
            <h5><img src={state.user.avatarurl} width="20%"></img></h5>
        </div>
    )
}

export default Profile