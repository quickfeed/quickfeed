import React, { useState } from "react"
import { EnrollmentStatus, EnrollmentStatusBadge, getCourseID, isPending, isTeacher, sortByField } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment, User } from "../../proto/ag/ag_pb"
import Search from "./Search"
import DynamicTable from "./DynamicTable"


const Members = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    const [func, setFunc] = useState("STATUS")
    const [descending, setDescending] = useState(false)
    const [edit, setEditing] = useState<boolean>(false)

    const sort = () => {
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
            actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.STUDENT })
        }
    }

    const pending = state.courseEnrollments[courseID].filter(enrollment => isPending(enrollment))

    const members = sortByField(state.courseEnrollments[courseID], [], sort(), descending).map((enrollment: Enrollment) => {
        const data: (string | JSX.Element)[] = []
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getName() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getEmail() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getStudentid() : "")
        data.push(enrollment.getLastactivitydate())
        data.push(enrollment.getTotalapproved().toString())

        if (isPending(enrollment)) {
            data.push(
                <div>
                    <i className="badge badge-primary" style={{ cursor: "pointer" }}
                        onClick={() => actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.STUDENT })}>
                        Accept
                    </i>
                    <i className="badge badge-danger clickable ml-1"
                        onClick={() => actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.NONE })}>
                        Reject
                    </i>
                </div>)
        } else {
            data.push(edit ? (
                <div>
                    <i className="badge badge-primary clickable"
                        onClick={() => actions.updateEnrollment({ enrollment: enrollment, status: isTeacher(enrollment) ? Enrollment.UserStatus.STUDENT : Enrollment.UserStatus.TEACHER })}>
                        {isTeacher(enrollment) ? "Demote" : "Promote"}
                    </i>
                    <i className="badge badge-danger clickable ml-1"
                        onClick={() => actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.NONE })}>
                        Reject
                    </i>
                </div>) :
                <i className={EnrollmentStatusBadge[enrollment.getStatus()]}>
                    {EnrollmentStatus[enrollment.getStatus()]}
                </i>
            )
        }
        return data
    })

    return (
        <div className='container'>
            <div className="row no-gutters pb-2">
                <div className="col-md-6">
                    <Search />
                </div>
                <div className="ml-auto">
                    <div className={edit ? "btn btn-sm btn-danger" : "btn btn-sm btn-primary"} onClick={() => setEditing(!edit)}>
                        {edit ? "Cancel" : "Edit"}
                    </div>
                </div>
                {pending.length > 0 ?
                    <div style={{ marginLeft: "10px" }}>
                        <button className="btn btn-success float-right" onClick={() => approveAll()}>
                            Approve All
                        </button>
                    </div> : null}
            </div>

            <div>
                <DynamicTable header={["Name", "Email", { value: "Student ID", onClick: () => { setFunc("ID"); setDescending(!descending) } }, "Activity", "Approved", { value: "Role", onClick: () => { setFunc("STATUS"); setDescending(!descending) } }]} data={members} />
            </div>
        </div>
    )
}


export default Members
