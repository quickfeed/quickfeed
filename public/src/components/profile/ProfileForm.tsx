import React, { Dispatch, SetStateAction, useCallback, useState } from "react"
import { hasEnrollment } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import FormInput from "../forms/FormInput"
import { useHistory } from "react-router"
import { clone } from "@bufbuild/protobuf"
import { UserSchema } from "../../../proto/qf/types_pb"

const ProfileForm = ({ children, setEditing }: { children: React.ReactNode, setEditing: Dispatch<SetStateAction<boolean>> }) => {
    const state = useAppState()
    const actions = useActions().global
    const history = useHistory()

    // Create a copy of the user object, so that we can modify it without affecting the original object.
    const [user, setUser] = useState(clone(UserSchema, state.self))
    const [isValid, setIsValid] = useState(state.isValid)

    // Update the user object when user input changes, and update the state.
    const handleChange = useCallback((event: React.FormEvent<HTMLInputElement>) => {
        const { name, value } = event.currentTarget
        switch (name) {
            case "name":
                user.Name = value
                break
            case "email":
                user.Email = value
                break
            case "studentid":
                user.StudentID = value
                break
        }
        setUser(user)
        if (user.Name !== "" && user.Email !== "" && user.StudentID !== "") {
            setIsValid(true)
        } else {
            setIsValid(false)
        }
    }, [user])


    // Sends the updated user object to the server on submit.
    const submitHandler = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        actions.updateUser(user)
        // Disable editing after submission
        setEditing(false)
        if (!hasEnrollment(state.enrollments)) {
            history.push("/courses")
        }
    }

    return (
        <div>
            {!isValid ? children : null}
            <form className="form-group" onSubmit={submitHandler}>
                <FormInput prepend="Name" name="name" defaultValue={user.Name} onChange={handleChange} />
                <FormInput prepend="Email" name="email" defaultValue={user.Email} onChange={handleChange} type="email" />
                <FormInput prepend="Student ID" name="studentid" defaultValue={user.StudentID} onChange={handleChange} type="number" />
                <div className="col input-group mb-3">
                    <input className="btn btn-primary" disabled={!isValid} type="submit" value="Save" style={{ marginTop: "20px" }} />
                </div>
            </form>
        </div>
    )
}

export default ProfileForm
