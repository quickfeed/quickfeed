import React from "react"
import { useAppState } from "../overmind"
import Groups from "../components/Groups"
import GroupComponent from "../components/group/Group"
import GroupForm from "../components/group/GroupForm"
import { useCourseID } from "../hooks/useCourseID"


const GroupPage = () => {
    const state = useAppState()
    const courseID = useCourseID()

    if (state.isTeacher) {
        return <Groups />
    }

    if (!state.hasGroup(courseID.toString())) {
        return <GroupForm />
    }
    return <GroupComponent />
}

export default GroupPage
