import React, { useEffect, useState } from "react"
import { Enrollment, User } from "../../../proto/ag/ag_pb"
import { sortByField } from "../../Helpers"
import { useOvermind } from "../../overmind"
import { CourseGroup } from "../../overmind/state"




const CreateGroup = (props: {courseID: number}) => {
    const {state, actions} = useOvermind()
    const [users, setUsers] = useState<number[]>([])
    const [name, setName] = useState<string>("Group Name")
    const [query, setQuery] = useState<string>("")

    useEffect(() => {
        if (!users.includes(state.self.getId())) {
            actions.getEnrollmentsByCourse({courseID: props.courseID, statuses: [Enrollment.UserStatus.STUDENT]})
            users.push(state.self.getId())
            setUsers([...users])
        }
        console.log(users)
    })


    const check = (query: string, user?: User): boolean => {
        if (user) {
            return !user.getName().toLowerCase().includes(query.toLowerCase())
        } 
        return true
    }

    const updateGroupUsers = (user?: User) => {
        if (user) {
            if (user.getId() === state.self.getId()) { return }
            if (!users.includes(user.getId())) {
                users.push(user.getId())
                setUsers([...users])
            }
            else {
                setUsers(users.filter(e => e != user.getId()))
            }
        }
    }

    const updateGroupName = (input: string) => {
        let n = input.length == 0 ? "Group Name" : input
        setName(n)
    }

    return(
        <div className="container box"> 
            <div className="row">
                <div className="card well col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>
                        Students
                    </div>
                    <input onKeyUp={e => setQuery(e.currentTarget.value)} placeholder={"Search"}></input>

                    <ul className="list-group list-group-flush">
                        {sortByField(state.courseEnrollments[props.courseID], [Enrollment.prototype.getUser], User.prototype.getName).map((enrollment: Enrollment) => {
                            let user = enrollment.getUser()
                            if (user && user.getId() !== state.self.getId()) {

                                if (users.includes(user.getId())) { 
                                    return null 
                                }

                                return (
                                    <li hidden={check(query, user) || (user.getId() in users)} key={user.getId()} className="list-group-item">
                                        {user.getName()} 
                                        <i style={{float: "right", cursor:"pointer"}} onClick={() => updateGroupUsers(user)}>+</i>
                                    </li>
                                )
                            }
                        })} 
                    </ul>
                </div>
                
                <div className='col'>
                    <div className="card well col-md-offset-2" >
                        <div className="card-header " style={{textAlign: "center"}}>{name}</div>
                        <input placeholder={"Group Name:"} onKeyUp={e => updateGroupName(e.currentTarget.value)}></input>
                        {state.courseEnrollments[props.courseID].map(enrollment => {
                                let user = enrollment.getUser() ?? new User
                                if (users.includes(user.getId())) {
                                    return (
                                        <li key={enrollment.getId()} className="list-group-item">
                                            <img id="group-image" src={enrollment.getUser()?.getAvatarurl()}></img>
                                            {enrollment.getUser()?.getName()} 
                                            <i style={{float: "right", cursor:"pointer"}} onClick={() => updateGroupUsers(enrollment.getUser())}>-</i>
                                        </li>
                                        )
                                    }
                                return null
                            })} 
                        <div className="btn btn-primary" onClick={() => actions.createGroup({courseID: props.courseID, users: users, name: name})}> Create Group </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default CreateGroup