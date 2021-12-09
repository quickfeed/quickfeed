import React, { Dispatch, SetStateAction, useState } from "react"
import { useActions, useAppState } from "../../overmind"
import { User } from "../../../proto/ag/ag_pb"
import { json } from "overmind"
import FormInput from "./FormInput"

export const UserProfileForm = ({setEditing}: {setEditing: Dispatch<SetStateAction<boolean>>}) : JSX.Element => {
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
        actions.updateUser(user)
        // Flip back to the uneditable view
        setEditing(false)
    }

    return ( 
        <div className="box">
            <form className="form-group" onSubmit={e => {e.preventDefault(); submitHandler()}}>
                <FormInput prepend="Name" name="name" defaultValue={user.getName()} onChange={handleChange} />
                <FormInput prepend="Email" name="email" defaultValue={user.getEmail()} onChange={handleChange} />
                <FormInput prepend="Student ID" name="studentid" defaultValue={user.getStudentid()} onChange={handleChange} />
                
                <input className="btn btn-primary" type="submit" value="Save" style={{marginTop:"20px"}}/>
            </form>
        </div>
    )
}

export default UserProfileForm