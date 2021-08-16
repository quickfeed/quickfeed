import React, { Dispatch, SetStateAction, useState } from "react"
import { useActions, useAppState } from "../../overmind"
import { User } from "../../../proto/ag/ag_pb"
import { json } from "overmind"

export const UserProfileForm = ({setEditing}: {setEditing: Dispatch<SetStateAction<boolean>>;}) => {
    const state = useAppState()
    const actions = useActions()

    const [user, setUser] = useState<User>(json(state.self))

    // Updates local user state on change in an input field
    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const { name, value } = event.currentTarget
        switch (name) {
            case "name":
                user.setName(value)
                break
            case "email":
                user.setEmail(value)
                break
            case "studentid":
                user.setStudentid(value)
                break
        }
        setUser(user)
    }
    

    // Sends off the edited (or not) information to the server. ((Could change actions.changeUser to take (username, email, studentid) as args and create the new user object in the action, not sure what's best))
    const submitHandler = () => {
        actions.changeUser(user)
        // Flip back to the uneditable view
        setEditing(false)
    }

    return ( 
        <div className="box">
            <form className="form-group well" style={{width: "400px"}} onSubmit={e => {e.preventDefault(); submitHandler()}}>
                <label htmlFor={"name"}>Name</label>
                <input className="form-control" name="name" type="text" defaultValue={user.getName()} onChange={handleChange} />
                
                <label htmlFor={"email"}>Email</label>
                <input className="form-control" name="email" type="text" defaultValue={user.getEmail()} onChange={handleChange} />
                
                <label htmlFor={"studentid"}>Student ID</label>
                <input className="form-control" name="studentid" type="text" defaultValue={user.getStudentid()} onChange={handleChange} />
                
                <input className="btn btn-primary" type="submit" value="Save" style={{marginTop:"20px"}}/>
            </form>
        </div>
    )
}

export default UserProfileForm