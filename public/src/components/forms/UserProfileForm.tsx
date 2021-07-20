import React, { Dispatch, SetStateAction, useState } from "react"
import { useOvermind } from "../../overmind"
import { User } from "../../../proto/ag/ag_pb"

interface IProps {
    editing: boolean;
    setEditing: Dispatch<SetStateAction<boolean>>;
  }

export const UserProfileForm = (props: IProps) => {
    const {state, actions} = useOvermind()

    const [user, setUser] = useState({'name': state.self.getName(), 'email': state.self.getEmail(), 'studentid': state.self.getStudentid()})

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
        changedUser.setId(state.self.getId())
        changedUser.setName(user.name)
        changedUser.setEmail(user.email)
        changedUser.setStudentid(user.studentid.toString())
        changedUser.setIsadmin(state.self.getIsadmin())
        actions.changeUser(changedUser)
        // Flip back to the uneditable view
        props.setEditing(false)
    }

    return ( 
        <div className="box">
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

export default UserProfileForm