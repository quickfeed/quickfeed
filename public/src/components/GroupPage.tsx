import React, { useEffect } from "react"
import { useOvermind } from "../overmind"
import Group from "./group/Group"
import CreateGroup from "./group/CreateGroup"
import { getCourseID } from "../Helpers"


export const GroupPage = () => {
    const {state, actions} = useOvermind()
    const courseID = getCourseID()

    useEffect(() => {
        console.log(courseID)
        actions.getGroupByUserAndCourse(courseID)
    })

    if (!state.userGroup[courseID]) {
        return <CreateGroup courseID={courseID} />    
    }
    return <Group courseID={courseID} />
    
}

export default GroupPage