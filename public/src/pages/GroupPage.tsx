import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import CreateGroup from "../components/group/CreateGroup"
import { getCourseID, isTeacher } from "../Helpers"
import Groups from "../components/Groups"
import GroupComponent from "../components/group/Group"


export const GroupPage = () => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!isTeacher(state.enrollmentsByCourseId[courseID])) {
            actions.getGroupByUserAndCourse(courseID)
        }
    })

    if (isTeacher(state.enrollmentsByCourseId[courseID])) {
        return <Groups />
    }

    if (!state.userGroup[courseID]) {
        return <CreateGroup />    
    }
    return <GroupComponent />
    
}

export default GroupPage