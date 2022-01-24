import React, { useEffect } from "react"
import { useActions, useAppState } from "../overmind"
import { getCourseID } from "../Helpers"
import Groups from "../components/Groups"
import GroupComponent from "../components/group/Group"
import GroupForm from "../components/group/GroupForm"


const GroupPage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    useEffect(() => {
        if (!state.isTeacher) {
            actions.getGroupByUserAndCourse(courseID)
        }
    }, [])

    if (state.isTeacher) {
        return <Groups />
    }

    if (!state.userGroup[courseID]) {
        return <GroupForm />
    }
    return <GroupComponent />
}

export default GroupPage
