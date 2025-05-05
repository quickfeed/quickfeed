import React from "react"
import { isPendingGroup } from "../../Helpers"
import { useAppState } from "../../overmind"
import { useCourseID } from "../../hooks/useCourseID"


const GroupComponent = () => {
    const state = useAppState()
    const courseID = useCourseID()

    const group = state.userGroup[courseID.toString()]

    const pendingIcon = isPendingGroup(group) ? <span className="badge badge-warning ml-2">Pending</span> : null
    const members = group.users.map(user =>
        <li key={user.ID.toString()} className="list-group-item">
            <img src={user.AvatarURL} style={{ width: "23px", marginRight: "10px", borderRadius: "50%" }} alt="" />
            {user.Name}
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
