import React from "react"
import { useAppState } from "../overmind"
import { getCourseID } from "../Helpers"
import Groups from "../components/Groups"
import GroupComponent from "../components/group/Group"
import GroupForm from "../components/group/GroupForm"


const GroupPage = (): JSX.Element => {
    const state = useAppState()
    const courseID = getCourseID()

    if (state.isTeacher) {
        return <Groups />
    }

    if (!state.hasGroup(Number(courseID))) {
        return <GroupForm />
    }
    return <GroupComponent />
}

export default GroupPage
