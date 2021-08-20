import { json } from "overmind"
import React, { useEffect, useState } from "react"
import { Group } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import GroupForm from "./group/GroupForm"
import Search from "./Search"

/* Lists all groups for a given course. */
export const Groups = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()


    const [editing, setEditing] = useState<Group>()

    useEffect(() => {
        actions.getGroupsByCourse(courseID)
    }, [state.groups, state.query])

    const groupSearch = (group: Group) => {
        // Show all groups if query is empty
        if (state.query.length == 0) {
            return false
        }

        // Show group if group name includes query
        if (group.getName().toLowerCase().includes(state.query)) { 
            return false
        }
        
        // Show group if any group user includes query
        for (const user of group.getUsersList()) {
            if (user.getName().toLowerCase().includes(state.query)) {
                return false
            }
        }
        // Hide group if none of the above include query
        return true
    }

    const deleteGroup = (group: Group) => {
        if (confirm("Deleting a group is an irreversible action. Are you sure?")) {
            actions.deleteGroup(group)
        }

    }

    const updateGroupStatus = (group: Group, status: Group.GroupStatus) => {
        actions.updateGroupStatus({group, status})
    }

    const GroupButtons = ({group}: {group: Group}) => {
        if (group.getStatus() === Group.GroupStatus.PENDING) {
            return (
                <li className="list-group-item">
                        <span onClick={() => setEditing(group)}>EDIT</span>
                        <span onClick={() => updateGroupStatus(group, Group.GroupStatus.APPROVED)} className="badge badge-primary float-right">Approve</span>
                        <span onClick={() => deleteGroup(group)} className="badge badge-danger float-right">Delete</span>
                </li>
            )
        }
        return <li className="list-group-item"><span onClick={() => setEditing(group)}>EDIT</span></li>
    }

    const GroupList = ({group}: {group: Group}) => {
        const style = {width: "40px", borderRadius: "50%", marginRight: "10px"}
        const classname = group.getStatus() == Group.GroupStatus.APPROVED ? "list-group-item active" : "list-group-item list-group-item-warning"
            return (
                <><ul key={group.getId()} hidden={groupSearch(group)} className="list-group list-group-flush">
                    <li className={classname}>
                        {group.getName()}
                        <span className="float-right badge badge-warning">{group.getStatus() == Group.GroupStatus.PENDING ? "Pending" : null}</span>
                    </li>
                    {// Populates the unordered list with list elements for every user in the group
                        group.getEnrollmentsList().map(enrol => 
                        <li key={enrol.getId()} className="list-group-item">
                            <img src={enrol.getUser()?.getAvatarurl()} style={style}></img>
                            {enrol.getUser()?.getName()}
                        </li>
                        )}
                    <GroupButtons group={group} />
                </ul></>
            )
    }

    // Generates JSX.Element array containing all groups for the course
    const PendingGroups = state.groups[courseID]?.filter(g => g.getStatus() == Group.GroupStatus.PENDING).map(group => {
        return <GroupList key={group.getId()} group={json(group)} />
    })

    const ApprovedGroups = state.groups[courseID]?.filter(g => g.getStatus() == Group.GroupStatus.APPROVED).map(group => {
        return <GroupList key={group.getId()} group={json(group)} />
    })

    if (editing) {
        return <div className="box"><GroupForm editGroup={editing} setGroup={setEditing} /></div>
    }

    return (
        <div className="box">
            <Search />
            <div className="row">
                <div className="col-sm-6">
                    {ApprovedGroups}
                </div>
                <div className="col-sm-6">
                    {PendingGroups}
                </div>
            </div>
        </div>
    )
}

export default Groups