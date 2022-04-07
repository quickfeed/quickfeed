import React, { useState } from "react"
import { Color, EnrollmentSort, EnrollmentStatus, EnrollmentStatusBadge, getCourseID, isPending, isTeacher, sortEnrollments } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment, User } from "../../proto/ag/ag_pb"
import Search from "./Search"
import DynamicTable, { Row } from "./DynamicTable"
import { json } from "overmind"
import DynamicButton from "./DynamicButton"
import { ButtonType } from "./admin/Button"


const Members = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    const [sortBy, setSortBy] = useState<EnrollmentSort>(EnrollmentSort.Status)
    const [descending, setDescending] = useState<boolean>(false)
    const [edit, setEditing] = useState<boolean>(false)
    const [accept, setAccept] = useState<boolean>(false)
    const [promote, setPromote] = useState<boolean>(false)

    const approveAll = () => {
        for (const enrollment of pending) {
            actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.STUDENT })
        }
    }


    const setSort = (sort: EnrollmentSort) => {
        if (sortBy === sort) {
            setDescending(!descending)
        }
        setSortBy(sort)
    }

    const pending = state.courseEnrollments[courseID].filter(enrollment => isPending(enrollment))

    const header: Row = [
        { value: "Name", onClick: () => setSort(EnrollmentSort.Name) },
        { value: "Email", onClick: () => setSort(EnrollmentSort.Email) },
        { value: "Student ID", onClick: () => setSort(EnrollmentSort.StudentID) },
        { value: "Activity", onClick: () => setSort(EnrollmentSort.Activity) },
        { value: "Approved", onClick: () => setSort(EnrollmentSort.Approved) },
        { value: "Slipdays", onClick: () => { setSort(EnrollmentSort.Slipdays) } },
        { value: "Role", onClick: () => { setSort(EnrollmentSort.Status) } },
    ]
    const members = sortEnrollments(json(state.courseEnrollments[courseID]), sortBy, descending).map(enrollment => {
        const data: Row = []
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getName() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getEmail() : "")
        data.push(enrollment.hasUser() ? (enrollment.getUser() as User).getStudentid() : "")
        data.push(enrollment.getLastactivitydate())
        data.push(enrollment.getTotalapproved().toString())
        data.push(enrollment.getSlipdaysremaining().toString())

        if (isPending(enrollment)) {
            data.push(
                <div>
                    <i className="badge badge-primary" style={{ cursor: "pointer" }}
                        onClick={() => { actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.STUDENT }); setAccept(!accept) }}>
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
                        onClick={() => { actions.updateEnrollment({ enrollment: enrollment, status: isTeacher(enrollment) ? Enrollment.UserStatus.STUDENT : Enrollment.UserStatus.TEACHER }); setPromote(!promote) }}>
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
                        <DynamicButton color={Color.GREEN} type={ButtonType.BUTTON} text="Approve All" onClick={() => actions.approvePendingEnrollments()} />
                    </div> : null}
            </div>

            <div>
                <DynamicTable header={header} data={members} />
            </div>
        </div>
    )
}

export default Members
