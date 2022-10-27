import React, { useEffect, useState } from "react"
import { Enrollment, Enrollment_UserStatus, Group } from "../../../gen/qf/types_pb"
import { getCourseID, hasTeacher, isApprovedGroup, isHidden, isPending, isStudent } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Search from "../Search"


const GroupForm = (): JSX.Element | null => {
    const state = useAppState()
    const actions = useActions()

    const [query, setQuery] = useState<string>("")
    const courseID = getCourseID()

    const group = state.activeGroup
    useEffect(() => {
        if (isStudent(state.enrollmentsByCourseID[courseID.toString()])) {
            actions.setActiveGroup(new Group())
            actions.updateGroupUsers(state.self.clone())
        }
        return () => {
            actions.setActiveGroup(null)
        }
    }, [])
    if (!group) {
        return null
    }
    const userIds = group.users.map(user => user.ID)

    const search = (enrollment: Enrollment): boolean => {
        if (userIds.includes(enrollment.userID) || enrollment.group && enrollment.groupID != group.ID) {
            return true
        }
        if (enrollment.user) {
            return isHidden(enrollment.user.name, query)
        }
        return false
    }

    const enrollments = state.courseEnrollments[courseID.toString()].map(enrollment => enrollment.clone())

    let sortedAndFilteredEnrollments: Enrollment[] = []
    if (hasTeacher(state.status[courseID.toString()])) {
        // If the current user has teacher status in the course, filter all enrollments where the user is not a teacher and is not in a group.
        sortedAndFilteredEnrollments = enrollments.filter(enrollment => enrollment.status == Enrollment_UserStatus.TEACHER && enrollment.groupID == BigInt(0))
    } else {
        // Filter out enrollments where the user is not a student, or the user is already in a group
        sortedAndFilteredEnrollments = enrollments.filter(enrollment => enrollment.status == Enrollment_UserStatus.STUDENT && enrollment.groupID == BigInt(0))
    }
    // Sort by name
    sortedAndFilteredEnrollments.sort((a, b) => (a.user?.name ?? "").localeCompare((b.user?.name ?? "")))

    const AvailableUser = ({ enrollment }: { enrollment: Enrollment }) => {
        const id = enrollment.userID
        if (isPending(enrollment)) {
            return null
        }
        if (id !== state.self.ID && !userIds.includes(id)) {
            return (
                <li hidden={search(enrollment)} key={id.toString()} className="list-group-item">
                    {enrollment.user?.name}
                    <i className="badge-pill badge-success ml-2 clickable float-right" onClick={() => actions.updateGroupUsers(enrollment.user)}>+</i>
                </li>
            )
        }
        return null
    }

    const groupMembers = group.users.map(user => {
        return (
            <li key={user.ID.toString()} className="list-group-item">
                <img id="group-image" src={user.avatarURL} alt="" />
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
                        {sortedAndFilteredEnrollments.map((enrollment, index) => {
                            return <AvailableUser key={index} enrollment={enrollment} />
                        })}
                    </ul>
                </div>

                <div className='col'>
                    <div className="card well col-md-offset-2" >
                        {GroupNameBanner}
                        {GroupNameInput}
                        {groupMembers}
                        {group && group.ID ?
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
