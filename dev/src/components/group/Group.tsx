import React from "react"
import { Group } from "../../../proto/ag/ag_pb"
import { getCourseID } from "../../Helpers"
import { useAppState } from "../../overmind"

const GroupComponent = (): JSX.Element => {
    const state = useAppState()
    const courseID = getCourseID()

    const GroupList = (group: Group) => {
        const elements: JSX.Element[] = []
        elements.push(<li className="list-group-item active">{state.userGroup[courseID].getName()}</li>)
        for (const user of group.getUsersList()) {
            elements.push(
                <li key={user.getId()} className="list-group-item">
                    <img src={user.getAvatarurl()} style={{width: "23px", marginRight: "10px", borderRadius: "50%"}}></img>
                    {user.getName()} 
                </li>)
        }
        return elements
    }

    return (
        <div className="box">
            {GroupList(state.userGroup[courseID]) }
        </div>
    )
}
export default GroupComponent