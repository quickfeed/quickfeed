import React, { useEffect, useState } from "react"
import { Enrollment, User } from "../../../proto/ag/ag_pb"
import { getCourseID, sortByField } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Search from "../Search"




const CreateGroup = () => {
    const state = useAppState()
    const actions = useActions()
    const [users, setUsers] = useState<number[]>([])
    const [name, setName] = useState<string>("Group Name")
    const courseID = getCourseID()

    useEffect(() => {
        if (!users.includes(state.self.getId()) && state.enrollmentsByCourseId[courseID].getStatus() != Enrollment.UserStatus.TEACHER) {
            actions.getEnrollmentsByCourse({courseID: courseID, statuses: [Enrollment.UserStatus.STUDENT]})
            users.push(state.self.getId())
            setUsers([...users])
        }
    })


    const check = (user?: User): boolean => {
        if (user) {
            return !user.getName().toLowerCase().includes(state.query)
        } 
        return true
    }

    const updateGroupUsers = (id: number) => {
        if (id === state.self.getId()) { 
            return 
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
        const name = input.length == 0 ? "Group Name" : input
        setName(name)
    }

    return(
        <div className="container box"> 
            <div className="row">
                <div className="card well col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>
                        Students
                    </div>
                    <Search placeholder={"Search"} />

                    <ul className="list-group list-group-flush">
                        {sortByField(state.courseEnrollments[courseID], [Enrollment.prototype.getUser], User.prototype.getName).map((enrollment: Enrollment) => {
                            let id = enrollment.getUserid()
                            if (id !== state.self.getId()) {

                                if (users.includes(id)) { 
                                    return null 
                                }

                                return (
                                    <li hidden={check(enrollment.getUser()) || (id in users) || enrollment.getGroupid() > 0} key={id} className="list-group-item">
                                        {enrollment.getUser()?.getName()} 
                                        <i style={{float: "right", cursor:"pointer"}} onClick={() => updateGroupUsers(id)}>+</i>
                                    </li>
                                )
                            }
                        })} 
                    </ul>
                </div>
                
                <div className='col'>
                    <div className="card well col-md-offset-2" >
                        <div className="card-header" style={{textAlign: "center"}}>{name}</div>
                        <input placeholder={"Group Name:"} onKeyUp={e => updateGroupName(e.currentTarget.value)}></input>
                        {state.courseEnrollments[courseID].map(enrollment => {
                                const id = enrollment.getUserid()
                                if (users.includes(id)) {
                                    return (
                                        <li key={enrollment.getId()} className="list-group-item">
                                            <img id="group-image" src={enrollment.getUser()?.getAvatarurl()}></img>
                                            {enrollment.getUser()?.getName()} 
                                            <i style={{float: "right", cursor:"pointer"}} onClick={() => updateGroupUsers(id)}>-</i>
                                        </li>
                                        )
                                    }
                                return null
                            })} 
                        <div className="btn btn-primary" onClick={() => actions.createGroup({courseID: courseID, users: users, name: name})}> Create Group </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default CreateGroup