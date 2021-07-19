import React, { useState } from "react"
import { useEffect } from "react"
import { Redirect, RouteComponentProps } from "react-router-dom"
import { EnrollmentStatus, sortByField } from "../Helpers"
import { useOvermind } from "../overmind"
import { Enrollment } from "../../proto/ag/ag_pb"


export const Members = (props: RouteComponentProps<{id?: string | undefined}>) => {
    const {state, actions} = useOvermind()
    let courseID = Number(props.match.params.id)

    const [func, setFunc] = useState("STATUS")
    const [descending, setDescending] = useState(true)
    useEffect(() => {

    }, [func, setFunc])

    const sort = (): Function => {
        switch (func) {
            case "STATUS":
                return Enrollment.prototype.getStatus
            case "ID":
                return Enrollment.prototype.getId
            default:
                return Enrollment.prototype.getStatus
        }

    }

    if (courseID && state.enrollmentsByCourseId[courseID].getStatus() === Enrollment.UserStatus.TEACHER) {
        const pending = state.courseEnrollments[courseID].filter(enrollment => enrollment.getStatus() === Enrollment.UserStatus.PENDING)
        
        return (
            <div className='row '>
                {pending.length > 0 ?
                <div className="card well  col-md-offset-2">
                    <div className="card-header" style={{textAlign: "center"}}>Pending</div>
                        <ul className="list-group list-group-flush">
  
                        {pending.map(user => {
                            if (user.getStatus() === Enrollment.UserStatus.PENDING) {
                                return (
                                    <li key={user.getUserid()} className={"list-group-item" }>{user.getUser()?.getName()} 
                                        <span className={"badge badge-primary"} onClick={() => actions.updateEnrollment({enrollment: user, status: Enrollment.UserStatus.STUDENT})}>
                                            Approve
                                        </span>
                                    </li>
                                )
                            }
                        })} 
                        </ul>
                </div>
                : null}
                <div className="col">
                    <div className="card well  col-md-offset-2">
                    <select onChange={(e) => setFunc(e.target.value)}>
                        <option value="NAME">Name</option>
                        <option selected value="STATUS">Status</option>
                        <option value="ID">ID</option>
                    </select>
                    Descending<input type={"checkbox"} checked={descending} onChange={(e) => setDescending(e.target.checked)}></input>
                        <div className="card-header" style={{textAlign: "center"}}>Members</div>
                            <ul className="list-group list-group-flush">
                                {sortByField(state.courseEnrollments[courseID], [], sort(), descending).map((user: Enrollment) => {
                                return (
                                    <li key={user.getUserid()} className={"list-group-item" }>
                                        {user.getUser()?.getName()} ({user.getUser()?.getStudentid()})
                                        <i style={{float: "right"}} 
                                            className={"badge badge-" + (user.getStatus() === 2 ? "primary" : user.getStatus() === 3 ? "danger" : "info")}>
                                                {EnrollmentStatus[user.getStatus()]}
                                        </i>
                                    </li>
                                )
                                })} 
                            </ul>
                    </div>
                </div>
            </div>
        )
    }
    return (<Redirect to="/" />)
}

export default Members