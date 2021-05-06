import React from "react"
import { Redirect } from "react-router"
import { useOvermind } from "../overmind"
import CourseCreationForm from "./forms/CourseCreationForm"



export const Admin = () => {
    const {state} = useOvermind()

    // Ideas: Statistics, Create Course, Promote Users

    if (state.user.isadmin) {
        return (
            
            <div className="box">
                <CourseCreationForm></CourseCreationForm>
            </div>
        )
    }
    return (
        <Redirect to="" />
    )
}

export default Admin