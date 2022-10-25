import React from "react"
import { Group, Group_GroupStatus } from "../../gen/qf/types_pb"
import { getCourseID, hasEnrollments, isApprovedGroup, isPendingGroup } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import GroupForm from "./group/GroupForm"
import Search from "./Search"


/* Lists all groups for a given course. */
const Groups = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    const groupSearch = (group: Group) => {
        // Show all groups if query is empty
        if (state.query.length == 0) {
            return false
        }

        // Show group if group name includes query
        if (group.name.toLowerCase().includes(state.query)) {
            return false
        }

        // Show group if any group user includes query
        for (const user of group.users) {
            if (user.name.toLowerCase().includes(state.query)) {
                return false
            }
        }
        // Hide group if none of the above include query
        return true
    }

    const updateGroupStatus = (group: Group, status: Group_GroupStatus) => {
        actions.updateGroupStatus({ group, status })
    }

    const GroupButtons = ({ group }: { group: Group }) => {
        const buttons: JSX.Element[] = []
        if (isPendingGroup(group)) {
            buttons.push(<span onClick={() => updateGroupStatus(group, Group_GroupStatus.APPROVED)} className="badge badge-primary clickable">Approve</span>)
        }
        buttons.push(<span className="badge badge-info clickable ml-2" onClick={() => actions.setActiveGroup(group)}>Edit</span>)
        buttons.push(<span onClick={() => actions.deleteGroup(group)} className="badge badge-danger clickable ml-2">Delete</span>)

        return <td>{buttons}</td>
    }

    const GroupMembers = ({ group }: { group: Group }) => {
        if (!hasEnrollments(group)) {
            return <td>No members</td>
        }

        const members = group.enrollments.map((enrollment, index) => {
            return (
                <span key={enrollment.ID.toString()} className="inline-block">
                    <a href={`https://github.com/${enrollment.user?.login}`} target="_blank" rel="noopener noreferrer">{enrollment.user?.name}</a>
                    {index >= group.enrollments.length - 1 ? "" : ", "}
                </span>
            )
        })
        return <td>{members}</td>
    }

    const GroupRow = ({ group }: { group: Group }) => {
        return (
            <tr hidden={groupSearch(group)}>
                <td key={group.ID.toString()}>
                    {group.name}
                    <span className="badge badge-warning ml-2">{isPendingGroup(group) ? "Pending" : null}</span>
                </td>
                <GroupMembers group={group} />
                <GroupButtons group={group} />
            </tr>
        )
    }

    // Generates JSX.Element array containing all groups for the course
    const PendingGroups = state.groups[courseID.toString()]?.filter(group => isPendingGroup(group)).map(group => {
        return <GroupRow key={group.ID.toString()} group={group} />
    })

    const ApprovedGroups = state.groups[courseID.toString()]?.filter(group => isApprovedGroup(group)).map(group => {
        return <GroupRow key={group.ID.toString()} group={group} />
    })

    // If a group is active (being edited), show the group form
    if (state.activeGroup) {
        return <GroupForm />
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
