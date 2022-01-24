import { json } from "overmind/lib/utils"
import React, { useEffect, useState } from "react"
import { Enrollment, Group, User } from "../../../proto/ag/ag_pb"
import { getCourseID, isApprovedGroup, isHidden, isStudent, sortByField } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Search from "../Search"


const GroupForm = ({ editGroup, setGroup }: { editGroup?: Group, setGroup?: React.Dispatch<React.SetStateAction<Group | undefined>> }): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const [users, setUsers] = useState<number[]>([])
    const [name, setName] = useState<string>("Group Name")
    const [query, setQuery] = useState<string>("")
    const courseID = getCourseID()

    const group = editGroup ? json(editGroup) : null

    useEffect(() => {
        if (!users.includes(state.self.getId()) && isStudent(state.enrollmentsByCourseID[courseID])) {
            // Add self to group if not teacher
            users.push(state.self.getId())
            setUsers([...users])
        }
        if (group) {
            users.splice(0, users.length)
            for (const user of group.getUsersList()) {
                users.push(user.getId())
            }
        }
        actions.getEnrollmentsByCourse({ courseID: courseID, statuses: [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER] })
    }, [group])


    const search = (enrollment: Enrollment): boolean => {
        if (enrollment.getUserid() in users || enrollment.hasGroup() && enrollment.getGroupid() != group?.getId()) {
            return true
        }
        if (enrollment.getUser()) {
            return isHidden((enrollment.getUser() as User).getName(), query)
        }
        return false
    }

    const indexOf = (users: User[], id: number): number => {
        for (let i = 0; i < users.length; i++) {
            if (users[i].getId() === id) {
                return i
            }
        }
        return -1
    }

    const updateGroupUsers = (id: number) => {
        if (id === state.self.getId()) {
            return
        }

        if (group) {
            const userList = group.getUsersList()
            const index = indexOf(userList, id)
            if (index >= 0) {
                // Remove user with id from group
                userList.splice(index, 1)
            } else {
                // User with id not found in group; add user
                const enrollment = state.courseEnrollments[courseID].find(enrollment => enrollment.getUserid() == id)
                const user = enrollment?.hasUser() ? enrollment.getUser() : null
                userList.push(user as User)
            }
            group.setUsersList(userList)
        }
        if (!users.includes(id)) {
            users.push(id)
            setUsers([...users])
        } else {
            setUsers(users.filter(e => e != id))
        }

    }

    const updateGroupName = (input: string) => {
        if (group) {
            group.setName(input)
        } else {
            const name = input.length == 0 ? "Group Name" : input
            setName(name)
        }
    }

    const sortedEnrollments = sortByField(state.courseEnrollments[courseID], [Enrollment.prototype.getUser], User.prototype.getName) as Enrollment[]

    const AvailableUser = ({ enrollment }: { enrollment: Enrollment }) => {
        const id = enrollment.getUserid()
        if (id !== state.self.getId() && !users.includes(id)) {
            return (
                <li hidden={search(enrollment)} key={id} className="list-group-item">
                    {enrollment.getUser()?.getName()}
                    <i className="badge-pill badge-success ml-2 clickable float-right" onClick={() => updateGroupUsers(id)}>+</i>
                </li>
            )
        }
        return null
    }

    const GroupMember = ({ enrollment }: { enrollment: Enrollment }) => {
        const id = enrollment.getUserid()
        if (users.includes(id)) {
            return (
                <li key={id} className="list-group-item">
                    <img id="group-image" src={enrollment.getUser()?.getAvatarurl()}></img>
                    {enrollment.getUser()?.getName()}
                    <i className="badge-pill badge-danger rounded-circle clickable float-right" onClick={() => updateGroupUsers(id)}>-</i>
                </li>
            )
        }
        return null
    }

    const GroupNameBanner = group
        ? <div className="card-header" style={{ textAlign: "center" }}>{group.getName()}</div>
        : <div className="card-header" style={{ textAlign: "center" }}>{name}</div>
    const GroupNameInput = group && isApprovedGroup(group)
        ? null
        : <input placeholder={"Group Name:"} onKeyUp={e => updateGroupName(e.currentTarget.value)}></input>

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
                        {state.courseEnrollments[courseID].map((enrollment, index) => {
                            return <GroupMember key={index} enrollment={enrollment} />
                        })}
                        {group && setGroup ?
                            <div className="row justify-content-md-center">
                                <div className="btn btn-primary ml-2" onClick={() => actions.updateGroup(group)}> Update </div>
                                <div className="btn btn-danger ml-2" onClick={() => setGroup(undefined)}> Cancel </div>
                            </div>
                            :
                            <div className="btn btn-primary" onClick={() => actions.createGroup({ courseID: courseID, users: users, name: name })}> Create Group </div>
                        }
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GroupForm
