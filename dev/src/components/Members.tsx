import React, { useState } from "react"
import { useEffect } from "react"
import { Redirect } from "react-router-dom"
import { EnrollmentStatus, getCourseID, isTeacher, sortByField } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment } from "../../proto/ag/ag_pb"


// TODO: Clean up 

export const Members = () => {
    const state = useAppState()
    const actions = useActions()
    let courseID = getCourseID()

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

    if (!isTeacher(state.enrollmentsByCourseId[courseID])) {
        return <Redirect to="/" />
    }

    const pending = state.courseEnrollments[courseID].filter(enrollment => enrollment.getStatus() === Enrollment.UserStatus.PENDING)
        
    return (
        <div className='row '>
                {pending.length > 0 ?
                <div className="col col-sm-4">
                    <div className="card-header" style={{textAlign: "center"}}>Pending</div>
                        <ul className="list-group list-group-flush">
  
                        {pending.map(user => {
                            if (user.getStatus() === Enrollment.UserStatus.PENDING) {
                                return (
                                    <li key={user.getUserid()} className={"list-group-item" }>{user.getUser()?.getName()} 
                                        <span className={"badge badge-primary float-right"} onClick={() => actions.updateEnrollment({enrollment: user, status: Enrollment.UserStatus.STUDENT})}>
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
                    <label htmlFor={"descending"}>Descending</label>
                    <input type={"checkbox"} name="descending" checked={descending} onChange={(e) => setDescending(e.target.checked)}></input>
                        <div className="card-header" style={{textAlign: "center"}}>Members</div>
                            <ul className="list-group list-group-flush">
                                {sortByField(state.courseEnrollments[courseID], [], sort(), descending).map((enrollment: Enrollment) => {
                                return (
                                    <li key={enrollment.getUserid()} className={"list-group-item" }>
                                        {enrollment.getUser()?.getName()} ({enrollment.getUser()?.getStudentid()})
                                        <i style={{float: "right"}} 
                                            className={"badge badge-" + (enrollment.getStatus() === 2 ? "primary" : enrollment.getStatus() === 3 ? "danger" : "info")}>
                                                {EnrollmentStatus[enrollment.getStatus()]}
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


export default Members
