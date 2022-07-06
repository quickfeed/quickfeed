import React, { useEffect, useState } from "react"
import { Enrollment, Group, User } from "../../../proto/qf/qf_pb"
import { Converter } from "../../convert"
import { getCourseID, isApprovedGroup, isHidden, isStudent, sortByField } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Search from "../Search"


const GroupForm = (): JSX.Element | null => {
    const state = useAppState()
    const actions = useActions()

    const [query, setQuery] = useState<string>("")
    const courseID = getCourseID()

    const group = state.activeGroup
    useEffect(() => {
        if (isStudent(state.enrollmentsByCourseID[courseID])) {
            actions.setActiveGroup(Converter.create<Group.AsObject>(Group))
            actions.updateGroupUsers(Converter.clone(state.self))
        }
        return () => {
            actions.setActiveGroup(null)
        }
    }, [])
    if (!group) {
        return null
    }
    const userIds = group.usersList.map(user => user.id)

    const search = (enrollment: Enrollment.AsObject): boolean => {
        if (enrollment.userid in userIds || enrollment.group && enrollment.groupid != group.id) {
            return true
        }
        if (enrollment.user) {
            return isHidden(enrollment.user.name, query)
        }
        return false
    }

    const enrollments = Converter.clone(state.courseEnrollments[courseID])

    const sortedEnrollments = sortByField(enrollments, [Enrollment.prototype.getUser], User.prototype.getName) as Enrollment.AsObject[]

    const AvailableUser = ({ enrollment }: { enrollment: Enrollment.AsObject }) => {
        const id = enrollment.userid
        if (id !== state.self.id && !userIds.includes(id)) {
            return (
                <li hidden={search(enrollment)} key={id} className="list-group-item">
                    {enrollment.user?.name}
                    <i className="badge-pill badge-success ml-2 clickable float-right" onClick={() => actions.updateGroupUsers(enrollment.user)}>+</i>
                </li>
            )
        }
        return null
    }

    const groupMembers = group.usersList.map(user => {
        return (
            <li key={user.id} className="list-group-item">
                <img id="group-image" src={user.avatarurl} alt="" />
                {user.name}
                <i className="badge-pill badge-danger rounded-circle clickable float-right" onClick={() => actions.updateGroupUsers(user)}>-</i>
            </li>
        )
    })

    const GroupNameBanner = <div className="card-header" style={{ textAlign: "center" }}>{group.name}</div>
    const GroupNameInput = group && isApprovedGroup(group)
        ? null
        : <input placeholder={"Group Name:"} onKeyUp={e => actions.updateGroupName(e.currentTarget.value)} />

    return (
        <div className="container">
            <div className="row">
                <div className="card well col-md-offset-2">
                    <div className="card-header" style={{ textAlign: "center" }}>
                        Students
                    </div>
                    <Search placeholder={"Search"} setQuery={setQuery} />

                    <ul className="list-group list-group-flush">
                        {sortedEnrollments.map((enrollment, index) => {
                            return <AvailableUser key={index} enrollment={enrollment} />
                        })}
                    </ul>
                </div>

                <div className='col'>
                    <div className="card well col-md-offset-2" >
                        {GroupNameBanner}
                        {GroupNameInput}
                        {groupMembers}
                        {group && group.id ?
                            <div className="row justify-content-md-center">
                                <div className="btn btn-primary ml-2" onClick={() => actions.updateGroup(group)}> Update </div>
                                <div className="btn btn-danger ml-2" onClick={() => actions.setActiveGroup(null)}> Cancel </div>
                            </div>
                            :
                            <div className="btn btn-primary" onClick={() => actions.createGroup({ courseID: courseID, users: userIds, name: group.name })}> Create Group </div>
                        }
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GroupForm
