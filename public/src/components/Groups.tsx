import React, { useCallback } from "react"
import { Group, Group_GroupStatus } from "../../proto/qf/types_pb"
import { Color, hasUsers, isApprovedGroup, isPendingGroup } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Button, { ButtonType } from "./admin/Button"
import DynamicButton from "./DynamicButton"
import GroupForm from "./group/GroupForm"
import Search from "./Search"
import { useCourseID } from "../hooks/useCourseID"


/* Lists all groups for a given course. */
const Groups = () => {
    const state = useAppState()
    const actions = useActions()
    const courseID = useCourseID()

    const groupSearch = (group: Group) => {
        // Show all groups if query is empty
        if (state.query.length === 0) {
            return false
        }

        // Show group if group name includes query
        if (group.name.toLowerCase().includes(state.query)) {
            return false
        }

        // Show group if any group user includes query
        for (const user of group.users) {
            if (user.Name.toLowerCase().includes(state.query)) {
                return false
            }
        }
        // Hide group if none of the above include query
        return true
    }

    const approveGroup = useCallback((group: Group) => () => actions.updateGroupStatus({ group, status: Group_GroupStatus.APPROVED }), [actions])
    const handleEditGroup = useCallback((group: Group) => () => actions.setActiveGroup(group), [actions])
    const handleDeleteGroup = useCallback((group: Group) => () => actions.deleteGroup(group), [actions])

    const GroupButtons = ({ group }: { group: Group }) => {
        const buttons: React.JSX.Element[] = []
        if (isPendingGroup(group)) {
            buttons.push(
                <DynamicButton
                    key={`approve${group.ID}`}
                    text="Approve"
                    color={Color.BLUE}
                    type={ButtonType.BADGE}
                    onClick={approveGroup(group)}
                />
            )
        }
        buttons.push(
            <Button
                key={`edit${group.ID}`}
                text="Edit"
                color={Color.YELLOW}
                type={ButtonType.BADGE}
                className="ml-2"
                onClick={handleEditGroup(group)}
            />
        )
        buttons.push(
            <DynamicButton
                key={`delete${group.ID}`}
                text="Delete"
                color={Color.RED}
                type={ButtonType.BADGE}
                className="ml-2"
                onClick={handleDeleteGroup(group)}
            />
        )

        return <td className="d-flex">{buttons}</td>
    }

    const GroupMembers = ({ group }: { group: Group }) => {
        if (!hasUsers(group)) {
            return <td>No members</td>
        }

        const members = group.users.map((user, index) => {
            return (
                <span key={user.ID.toString()} className="inline-block">
                    <a href={`https://github.com/${user.Login}`} target="_blank" rel="noopener noreferrer">{user.Name}</a>
                    {index >= group.users.length - 1 ? "" : ", "}
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
        return <GroupForm key={state.activeGroup.ID.toString()} />
    }

    const table = (
        <table className="table table-striped table-grp table-hover">
            <thead className="thead-dark">
                <tr>
                    <th>Name</th>
                    <th>Members</th>
                    <th>Manage</th>
                </tr>
            </thead>
            <tbody>
                {PendingGroups}
                {ApprovedGroups}
            </tbody>
        </table>
    )

    return (
        <div className="box">
            <div className="pb-2">
                <Search />
            </div>
            {table}
        </div>
    )
}

export default Groups
