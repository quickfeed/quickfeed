import React, { useEffect } from "react"
import { RouteComponentProps } from "react-router"
import { useOvermind } from "../overmind"
import Group from "./group/Group"
import CreateGroup from "./group/CreateGroup"


export const GroupPage = (props: RouteComponentProps<{id?: string | undefined}>) => {
    const {state, actions} = useOvermind()
    const courseID = Number(props.match.params.id)

    useEffect(() => {
        actions.getGroupByUserAndCourse(courseID)
    })

    if (!state.userGroup[courseID]) {
        return <CreateGroup courseID={courseID} />    
    }
    return <Group courseID={courseID} />
    
}

export default GroupPage