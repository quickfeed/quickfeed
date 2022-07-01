import React from "react"
import { getCourseID, isPendingGroup } from "../../Helpers"
import { useAppState } from "../../overmind"


const GroupComponent = (): JSX.Element => {
    const state = useAppState()
    const courseID = getCourseID()

    const group = state.userGroup[courseID]

    const pendingIcon = isPendingGroup(group) ? <span className="badge badge-warning ml-2">Pending</span> : null
    const members = group.usersList.map(user =>
        <li key={user.id} className="list-group-item">
            <img src={user.avatarurl} style={{ width: "23px", marginRight: "10px", borderRadius: "50%" }} />
            {user.name}
        </li>
    )

    return (
        <div>
            <li className="list-group-item active">{group.name}{pendingIcon}</li>
            {members}
        </div>
    )
}

export default GroupComponent
