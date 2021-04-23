import React from "react"
import { Redirect } from "react-router"
import { useOvermind } from "../overmind"
import CreateCourse from "./CreateCourse"




export const Admin = () => {
    const {state, actions, effects} = useOvermind()


    // Ideas: Statistics, Create Course, Promote Users

    if (state.user.isadmin) {
        return (
            
            <div className="box">
                <CreateCourse></CreateCourse>
            </div>
        )
    }
    return (
        <Redirect to="" />
    )
}

export default Admin