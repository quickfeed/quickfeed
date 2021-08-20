import { json } from "overmind/lib/utils"
import React, { useEffect, useState } from "react"
import { Enrollment, Group, User } from "../../../proto/ag/ag_pb"
import { getCourseID, sortByField } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Search from "../Search"




const GroupForm = ({editGroup, setGroup}: {editGroup?: Group, setGroup?: React.Dispatch<React.SetStateAction<Group | undefined>>}): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const [users, setUsers] = useState<number[]>([])
    const [name, setName] = useState<string>("Group Name")
    const [query, setQuery] = useState<string>("")
    const courseID = getCourseID()

    const group = editGroup ? json(editGroup) : null

    useEffect(() => {
        // Add self to group if not teacher
        if (!users.includes(state.self.getId()) && state.enrollmentsByCourseId[courseID].getStatus() != Enrollment.UserStatus.TEACHER) {
            users.push(state.self.getId())
            setUsers([...users])
        }
        if (group) {
            users.splice(0,100)
            for (const user of group.getUsersList()) {
                users.push(user.getId())
            }
        }
        actions.getEnrollmentsByCourse({courseID: courseID, statuses: [Enrollment.UserStatus.STUDENT, Enrollment.UserStatus.TEACHER]})
    }, [group])


    const search = (enrollment: Enrollment): boolean => {
        if (enrollment.getUserid() in users || enrollment.hasGroup() && enrollment.getGroupid() != group?.getId()) {
            return true
        }
        if (enrollment.getUser()) {
            return !enrollment.getUser()?.getName().toLowerCase().includes(query)
        } 
        return false
    }

    const updateGroupUsers = (id: number) => {
        if (id === state.self.getId()) { 
            return 
        }

        if (group) {
            let added = false
            for (const user of group.getUsersList()) {
                if (id == user.getId()) {
                    group.getUsersList().splice(group.getUsersList().indexOf(user), 1)
                    group.setUsersList(group.getUsersList())
                    added = true
                }
            }
            if (!added) {
                const enrollment = state.courseEnrollments[courseID].find(enrollment => enrollment.getUserid() == id)
                if (enrollment?.hasUser()) {
                    group.getUsersList().push(json(enrollment).getUser() as User)
                    group.setUsersList(group.getUsersList())
                }
            }

        }
        if (!users.includes(id)) {
            users.push(id)
            setUsers([...users])
        }
        else {
            setUsers(users.filter(e => e != id))
        }
        
    }

    const updateGroupName = (input: string) => {
        if (group) {
            group.setName(input)
        } 
        else {
            const name = input.length == 0 ? "Group Name" : input
            setName(name)
        }
    }

    const sortedEnrollments = sortByField(state.courseEnrollments[courseID], [Enrollment.prototype.getUser], User.prototype.getName) as Enrollment[]

    const FreeUser = ({enrollment}: {enrollment: Enrollment}) => {
        const id = enrollment.getUserid()
        if (id !== state.self.getId() && !users.includes(id)) {
            return (
                <li hidden={search(enrollment)} key={id} className="list-group-item">
                    {enrollment.getUser()?.getName()} 
                    <i style={{float: "right", cursor:"pointer"}} onClick={() => updateGroupUsers(id)}>+</i>
                </li>
            )
        }
        return null
    }

    const GroupMember = ({enrollment}: {enrollment: Enrollment}) => {
        const id = enrollment.getUserid()
        if (users.includes(id)) {
            return (
                <li key={id} className="list-group-item">
                    <img id="group-image" src={enrollment.getUser()?.getAvatarurl()}></img>
                        {enrollment.getUser()?.getName()} 
                    <i style={{float: "right", cursor:"pointer"}} onClick={() => updateGroupUsers(id)}>-</i>
                </li>
            )
        }
        return null
    }

    const GroupNameBanner = group ? <div className="card-header" style={{textAlign: "center"}}>{group.getName()}</div> : <div className="card-header" style={{textAlign: "center"}}>{name}</div>
    const GroupNameInput = group && group.getStatus() === Group.GroupStatus.APPROVED ? null : <input placeholder={"Group Name:"} onKeyUp={e => updateGroupName(e.currentTarget.value)}></input>
   
    return(
        <div className="container box"> 
            <div className="row">
                <div className="card well col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>
                        Students
                    </div>
                    <Search placeholder={"Search"} setQuery={setQuery} />

                    <ul className="list-group list-group-flush">
                        {sortedEnrollments.map((enrollment, index) => {
                            return <FreeUser key={index} enrollment={enrollment} />
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
                            <>
                            <div className="btn btn-primary" onClick={() => actions.updateGroup(group)}> Update </div>
                            <div className="btn btn-primary" onClick={() => setGroup(undefined)}> Cancel </div> 
                            </>
                            : 
                            <div className="btn btn-primary" onClick={() => actions.createGroup({courseID: courseID, users: users, name: name})}> Create Group </div> 
                        }
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GroupForm