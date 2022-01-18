import React, { Dispatch, SetStateAction, useMemo } from "react"
import { useActions, useAppState } from "../../overmind"
import { json } from "overmind"
import FormInput from "../forms/FormInput"
import { hasEnrollment } from "../../Helpers"
import { useHistory } from "react-router"


const ProfileForm = ({ children, setEditing }: { children: React.ReactNode, setEditing: Dispatch<SetStateAction<boolean>> }): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const history = useHistory()

    const signup = useMemo(() => !state.isValid, [state.isValid])

    // Create a copy of the user object, so that we can modify it without affecting the original object.
    const user = json(state.self)

    // Update the user object when user input changes, and update the state.
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
        actions.setSelf(user)
    }


    // Sends the updated user object to the server on submit.
    const submitHandler = () => {
        actions.updateUser(user)
        // Disable editing after submission
        setEditing(false)
        if (!hasEnrollment(state.enrollments)) {
            history.push("/courses")
        }
    }

    return (
        <div>
            {signup ? children : null}
            <form className="form-group" onSubmit={e => { e.preventDefault(); submitHandler() }}>
                <FormInput prepend="Name" name="name" defaultValue={user.getName()} onChange={handleChange} />
                <FormInput prepend="Email" name="email" defaultValue={user.getEmail()} onChange={handleChange} type="email" />
                <FormInput prepend="Student ID" name="studentid" defaultValue={user.getStudentid()} onChange={handleChange} type="number" />
                <div className="col input-group mb-3">
                    <input className="btn btn-primary" disabled={!state.isValid} type="submit" value="Save" style={{ marginTop: "20px" }} />
                </div>
            </form>
        </div>
    )
}

export default ProfileForm
