import React, { useState } from "react"
import { Group } from "../../proto/ag/ag_pb"
import { getCourseID, hasEnrollments, isApprovedGroup, isPendingGroup } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import GroupForm from "./group/GroupForm"
import Search from "./Search"


/* Lists all groups for a given course. */
const Groups = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()


    const [editing, setEditing] = useState<Group.AsObject>()

    const groupSearch = (group: Group.AsObject) => {
        // Show all groups if query is empty
        if (state.query.length == 0) {
            return false
        }

        // Show group if group name includes query
        if (group.name.toLowerCase().includes(state.query)) {
            return false
        }

        // Show group if any group user includes query
        for (const user of group.usersList) {
            if (user.name.toLowerCase().includes(state.query)) {
                return false
            }
        }
        // Hide group if none of the above include query
        return true
    }

    const updateGroupStatus = (group: Group.AsObject, status: Group.GroupStatus) => {
        actions.updateGroupStatus({ group, status })
    }

    const GroupButtons = ({ group }: { group: Group.AsObject }) => {
        if (isPendingGroup(group)) {
            return (
                <td>
                    <span onClick={() => updateGroupStatus(group, Group.GroupStatus.APPROVED)} className="badge badge-primary clickable">Approve</span>
                    <span className="badge badge-info clickable ml-2" onClick={() => setEditing(group)}>Edit</span>
                    <span onClick={() => actions.deleteGroup(group)} className="badge badge-danger clickable ml-2">Delete</span>
                </td>
            )
        }
        return <td><span className="badge badge-info clickable" onClick={() => setEditing(group)}>Edit</span></td>
    }

    const GroupRow = ({ group }: { group: Group.AsObject }) => {
        return (
            <>
                <tr hidden={groupSearch(group)}>
                    <th key={group.id}>
                        {group.name}
                        <span className="badge badge-warning ml-2">{isPendingGroup(group) ? "Pending" : null}</span>
                    </th>
                    <td>
                        <div>
                            {// Populates the unordered list with list elements for every user in the group
                                hasEnrollments(group) && group.enrollmentsList.map((enrol, index) =>
                                    <span key={enrol.id} className="inline-block">
                                        <a href={`https://github.com/${enrol.user?.login}`} target="_blank" rel="noopener noreferrer">{enrol.user?.name}</a>
                                        {index >= group.enrollmentsList.length - 1 ? "" : ", "}
                                    </span>
                                )}
                        </div>
                    </td>
                    <GroupButtons group={group} />
                </tr>
            </>
        )
    }

    // Generates JSX.Element array containing all groups for the course
    const PendingGroups = state.groups[courseID]?.filter(group => isPendingGroup(group)).map(group => {
        return <GroupRow key={group.id} group={group} />
    })

    const ApprovedGroups = state.groups[courseID]?.filter(group => isApprovedGroup(group)).map(group => {
        return <GroupRow key={group.id} group={group} />
    })

    if (editing) {
        return <GroupForm editGroup={editing} setGroup={setEditing} />
    }

    return (
        <div className="box">
            <div className="pb-2">
                <Search />
            </div>
            <table className="table table-striped table-grp table-hover">
                <thead className="thead-dark">
                    <th>Name</th>
                    <th>Members</th>
                    <th>Manage</th>
                </thead>
                <tbody>
                    {PendingGroups}
                    {ApprovedGroups}
                </tbody>
            </table>
        </div>
    )
}

export default Groups
