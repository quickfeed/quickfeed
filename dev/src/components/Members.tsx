import React, { useState } from "react"
import { useEffect } from "react"
import { Redirect } from "react-router-dom"
import { EnrollmentStatus, EnrollmentStatusBadge, getCourseID, isTeacher, sortByField } from "../Helpers"
import { useAppState, useActions, useGrpc } from "../overmind"
import { Enrollment, User } from "../../proto/ag/ag_pb"
import Search from "./Search"
import DynamicTable from "./DynamicTable"


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

    const approveAll = () => {
        for (const enrollment of pending) {
            actions.updateEnrollment({enrollment: enrollment, status: Enrollment.UserStatus.STUDENT})
        }
    }

    if (!isTeacher(state.enrollmentsByCourseId[courseID])) {
        return <Redirect to="/" />
    }

    const Sort = () => {
        return (
        <div className="input-group">
            <select defaultValue="STATUS" className="form-control" onChange={(e) => setFunc(e.target.value)}>
                <option value="NAME">Name</option>
                <option value="STATUS">Status</option>
                <option value="ID">ID</option>
            </select>
            <div className="form-check form-check-inline">
                <input className="form-check-input" type={"checkbox"} name="descending" checked={descending} onChange={(e) => setDescending(e.target.checked)}></input>
                <label className="form-check-label" htmlFor={"descending"}>Descending</label>
            </div>
        </div>
        )
    }

    const pending = state.courseEnrollments[courseID].filter(enrollment => enrollment.getStatus() === Enrollment.UserStatus.PENDING)

    const pendingMembers = pending.map(enrollment => {
        const data: (string | JSX.Element)[] = []
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getName() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getEmail() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getStudentid() : "")
        data.push(
            <div>
                <button 
                    className="btn btn-primary" 
                    onClick={() => actions.updateEnrollment({enrollment: enrollment, status: Enrollment.UserStatus.STUDENT})}
                >
                    Accept
                </button>
                <button 
                    className="btn btn-danger" 
                    onClick={() => {
                        if (confirm("WARNNG! Rejecting a student is irreversible. Are you sure?"))
                            actions.updateEnrollment({enrollment: enrollment, status: Enrollment.UserStatus.NONE}) 
                        }}
                >
                    Reject
                </button>
            </div>
        )
        return data
    })

    const members = sortByField(state.courseEnrollments[courseID], [], sort(), descending).map((enrollment: Enrollment) => {
        const data: (string | JSX.Element)[] = []
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getName() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getEmail() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getStudentid() : "")
        data.push(enrollment.getLastactivitydate())
        data.push(enrollment.getTotalapproved().toString())
        data.push(
            <i className={EnrollmentStatusBadge[enrollment.getStatus()]}>
                {EnrollmentStatus[enrollment.getStatus()]}
            </i>
        )
        return data
    })

    return (
        <div className='container'>
            <Search />
            <div>
                {pending.length > 0 ? <h3>Pending Members<button className="btn btn-success float-right" onClick={() => approveAll()}>Approve All</button></h3>  : null}
                <DynamicTable header={["Name", "Email", "Student ID", "Role"]} data={pendingMembers} />
            </div>
            <div>
                <Sort />
                <DynamicTable header={["Name", "Email", "Student ID", "Activity", "Approved", "Role"]} data={members} />
            </div>
        </div>
        )
    }


export default Members
