import React from "react"
import Groups from "../components/Groups"
import GroupView from "../components/group/Group"
import GroupForm from "../components/group/GroupForm"
import { useCourseID } from "../hooks/useCourseID"
import { useAppState } from "../overmind"


const GroupPage = () => {
    const state = useAppState()
    const courseID = useCourseID()

    if (state.isTeacher) {
        return <Groups />
    }

    if (!state.hasGroup(courseID.toString())) {
        return <GroupForm />
    }
    return <GroupView />
}

export default GroupPage
