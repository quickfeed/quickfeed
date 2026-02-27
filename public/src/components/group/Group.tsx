import React from "react"
import { isPendingGroup } from "../../Helpers"
import { useAppState } from "../../overmind"
import { useCourseID } from "../../hooks/useCourseID"
import Badge from "../Badge"


const GroupView = () => {
    const state = useAppState()
    const courseID = useCourseID()

    const group = state.userGroup[courseID.toString()]

    if (!group) {
        return (
            <div className="card bg-base-100 shadow-xl">
                <div className="card-body">
                    <h2 className="card-title">No group</h2>
                    <p>You are not in a group for this course.</p>
                </div>
            </div>
        )
    }

    const pendingBadge = isPendingGroup(group) ? <Badge className="ml-2" text="Pending" color="yellow" type="ghost" /> : null

    const members = group.users.map(user =>
        <div
            key={user.ID.toString()}
            className="flex items-center gap-3 p-3 hover:bg-base-200 rounded-lg transition-colors"
        >
            <div className="avatar">
                <div className="w-10 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2">
                    <img src={user.AvatarURL} alt={`${user.Name}'s avatar`} />
                </div>
            </div>
            <span className="font-medium">{user.Name}</span>
        </div>
    )

    return (
        <div className="card bg-base-100 shadow-xl">
            <div className="card-body p-0">
                <div className="flex items-center justify-between bg-primary text-primary-content px-6 py-4 rounded-t-lg">
                    <div className="flex items-center gap-2">
                        <i className="fa fa-users text-xl"></i>
                        <h2 className="card-title text-xl">{group.name}</h2>
                    </div>
                    {pendingBadge}
                </div>
                <div className="p-4 space-y-1">
                    {members.length > 0 ? (
                        members
                    ) : (
                        <div className="text-center py-8 text-base-content/60">
                            <i className="fa fa-user-slash text-3xl mb-2"></i>
                            <p>No members yet</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}

export default GroupView
