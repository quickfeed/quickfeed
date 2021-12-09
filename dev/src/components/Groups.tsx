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

    const updateGroupStatus = (group: Group, status: Group.GroupStatus) => {
        actions.updateGroupStatus({group, status})
    }

    const GroupButtons = ({group}: {group: Group}) => {
        if (group.getStatus() === Group.GroupStatus.PENDING) {
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

    const GroupList = ({group}: {group: Group}) => {
            return (
                <>
                    <tr hidden={groupSearch(group)}>
                        <th key={group.getId()}> 
                            {group.getName()}
                            <span className="badge badge-warning ml-2">{group.getStatus() == Group.GroupStatus.PENDING ? "Pending" : null}</span>
                        </th>
                        <td>
                            <div>
                            {// Populates the unordered list with list elements for every user in the group
                                group.getEnrollmentsList().map((enrol, index) => 
                                <span key={enrol.getId()} className="inline-block">
                                    {/**<img src={enrol.getUser()?.getAvatarurl()} style={style}></img>*/}
                                    <a href={`https://github.com/${enrol.getUser()?.getLogin()}`} target="_blank" rel="noreferrer">{enrol.getUser()?.getName()}</a>
                                    {index >= group.getEnrollmentsList().length - 1 ? "" : ", "}
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
    const PendingGroups = state.groups[courseID]?.filter(g => g.getStatus() == Group.GroupStatus.PENDING).map(group => {
        return <GroupList key={group.getId()} group={json(group)} />
    })

    const ApprovedGroups = state.groups[courseID]?.filter(g => g.getStatus() == Group.GroupStatus.APPROVED).map(group => {
        return <GroupList key={group.getId()} group={json(group)} />
    })

    if (editing) {
        return <GroupForm editGroup={editing} setGroup={setEditing} />
    }

    return (
        <div className="box">
            <div className="pb-2">
                <Search />
            </div>
            <table className="table table-striped table-grp">
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