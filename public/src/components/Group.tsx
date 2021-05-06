import React, { useEffect, useState } from "react"
import { RouteComponentProps } from "react-router"
import { sortByField } from "../Helpers"
import { useOvermind } from "../overmind"
import { CourseGroup } from "../overmind/state"
import { Enrollment, User } from "../proto/ag_pb"
import SubmissionsTable from "./SubmissionsTable"

/*  The Group component provides an interface to create a group for a course.
    
*/
export const Group = (props: RouteComponentProps<{id?: string | undefined}>) => {
    const {state, actions} = useOvermind()
    const courseID = Number(props.match.params.id)
    const [cg, setCg] = useState<CourseGroup>({enrollments: [], groupName: "Group Name", users: []})
    const self = state.enrollmentsByCourseId[courseID].getUser()

    useEffect(() => {
        if (cg.users.length == 0 && self) {
            actions.getGroupByUserAndCourse(courseID)
            actions.getEnrollmentsByCourse({courseID: courseID, statuses: [Enrollment.UserStatus.STUDENT]})
            cg.users.push(self)
            setCg(cg)
        }
    }, [state.courseGroup])

    const updateSearchState = (e: React.FormEvent<HTMLInputElement>) => {
        actions.updateSearch(e.currentTarget.value)
    }

    const updateGroupUsers = (user: User | undefined, remove: boolean, enrollmentId?: number) => {
        if (user && enrollmentId && !remove) {
            if (!cg.users.includes(user)){
                cg.enrollments.push(enrollmentId)
                cg.users.push(user)
                setCg(cg)
            }
        }
        if (user && remove) {
            if(cg.users.includes(user)) {
                cg.users = cg.users.filter(e => e.getId() != user.getId())
                setCg(cg)
            }
        }
        actions.updateCourseGroup(cg)
    }

    const updateGroupName = (event: React.FormEvent<HTMLInputElement>) => {
        if (event.currentTarget.value.length == 0) {
            cg.groupName = "Group Name"
        }
        else {
            cg.groupName = event.currentTarget.value
        }
        setCg(cg)
        actions.updateCourseGroup(cg)
    }



    if (state.userGroup[courseID]) {
        const group = state.userGroup[courseID]
        return (
            <React.Fragment>
            
            <div className="box">
                <div className="jumbotron">
                    <div className="centerblock container">
                        <h1>Fancy Group Name</h1>
                        {state.enrollmentsByCourseId[courseID].getCourse()?.getName()}
                    </div>
                </div>
                {group.getUsersList().map(user => 
                    <li key={user.getId()} className="list-group-item">
                                        <img src={user.getAvatarurl()} style={{width: "23px", marginRight: "10px", borderRadius: "50%"}}></img>
                            {user.getName()} 
                    </li>
                )}
                <br />
                <SubmissionsTable courseID={courseID} group={true} />
            </div>
            </React.Fragment>
        )
    }

    return(
        <div className="container">
            
            <div className='row '>
                <div className="card well col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>
                        Students
                    </div>
                    <input onKeyUp={updateSearchState} placeholder={"Search"}></input>
                    <ul className="list-group list-group-flush">
                        {sortByField(state.userSearch, [Enrollment.prototype.getUser], User.prototype.getName).map((user: Enrollment) => {
                            if (user.getUser()?.getName() !== state.user.name) {
                                return (
                                    <li key={user.getUserid()} className="list-group-item">
                                        {user.getUser()?.getName()} 
                                        <i style={{float: "right"}} onClick={() => updateGroupUsers(user.getUser(), false, user.getId())}>+</i>
                                    </li>
                                )
                            }
                        })} 
                        </ul>
                </div>
                
                <div className='col'>
                    <div className="card well  col-md-offset-2" >
                    <div className="card-header " style={{textAlign: "center"}}>{cg.groupName}</div>
                    <input placeholder={"Group Name:"} onKeyUp={e => updateGroupName(e)}></input>
                    {cg.users.map(user => {
                            return (
                                <li key={user.getId()} className="list-group-item">
                                    <img src={user.getAvatarurl()} style={{width: "23px", marginRight: "10px", borderRadius: "50%"}}></img>
                                    {user.getName()} 
                                    <i style={{float: "right"}} onClick={() => updateGroupUsers(user, true)}>-</i>
                                </li>
                                )
                        })} 
                    <div className={"btn btn-primary"} onClick={() => actions.createGroup(courseID)}> Create Group </div>
                    </div>
                </div>
            </div>
        </div>
        )
}

export default Group