import React, { useState } from "react"
import { Color, EnrollmentSort, EnrollmentStatus, EnrollmentStatusBadge, getCourseID, isPending, isTeacher, sortEnrollments } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment } from "../../proto/qf/qf_pb"
import Search from "./Search"
import DynamicTable, { Row } from "./DynamicTable"
import DynamicButton from "./DynamicButton"
import { ButtonType } from "./admin/Button"
import { Converter } from "../convert"

const Members = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()

    const [sortBy, setSortBy] = useState<EnrollmentSort>(EnrollmentSort.Status)
    const [descending, setDescending] = useState<boolean>(false)
    const [edit, setEditing] = useState<boolean>(false)


    const setSort = (sort: EnrollmentSort) => {
        if (sortBy === sort) {
            setDescending(!descending)
        }
        setSortBy(sort)
    }

    let enrollments: Enrollment.AsObject[] = []
    if (state.courseEnrollments[courseID]) {
        // Clone the enrollments so we can sort them
        enrollments = Converter.clone(state.courseEnrollments[courseID])
    }

    const pending = state.pendingEnrollments

    const header: Row = [
        { value: "Name", onClick: () => setSort(EnrollmentSort.Name) },
        { value: "Email", onClick: () => setSort(EnrollmentSort.Email) },
        { value: "Student ID", onClick: () => setSort(EnrollmentSort.StudentID) },
        { value: "Activity", onClick: () => setSort(EnrollmentSort.Activity) },
        { value: "Approved", onClick: () => setSort(EnrollmentSort.Approved) },
        { value: "Slipdays", onClick: () => { setSort(EnrollmentSort.Slipdays) } },
        { value: "Role", onClick: () => { setSort(EnrollmentSort.Status) } },
    ]
    const members = sortEnrollments(enrollments, sortBy, descending).map(enrollment => {
        const data: Row = []
        data.push(enrollment.user ? enrollment.user.name : "")
        data.push(enrollment.user ? enrollment.user.email : "")
        data.push(enrollment.user ? enrollment.user.studentid : "")
        data.push(enrollment.lastactivitydate)
        data.push(enrollment.totalapproved.toString())
        data.push(enrollment.slipdaysremaining.toString())

        if (isPending(enrollment)) {
            data.push(
                <div>
                    <i className="badge badge-primary" style={{ cursor: "pointer" }}
                        onClick={() => { actions.updateEnrollment({ enrollment: enrollment, status: Enrollment.UserStatus.STUDENT }) }}>
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
                <i className={EnrollmentStatusBadge[enrollment.status]}>
                    {EnrollmentStatus[enrollment.status]}
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
                {pending?.length > 0 ?
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
