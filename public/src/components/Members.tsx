import React, { useState } from "react"
import { Color, EnrollmentSort, EnrollmentStatus, EnrollmentStatusBadge, getCourseID, getFormattedTime, isPending, isTeacher, sortEnrollments } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment, Enrollment_UserStatus } from "../../proto/qf/types_pb"
import Search from "./Search"
import DynamicTable, { Row } from "./DynamicTable"
import DynamicButton from "./DynamicButton"
import { ButtonType } from "./admin/Button"

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

    let enrollments: Enrollment[] = []
    if (state.courseEnrollments[courseID.toString()]) {
        // Clone the enrollments so we can sort them
        enrollments = state.courseEnrollments[courseID.toString()].slice()
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
        data.push(enrollment.user ? enrollment.user.Name : "")
        data.push(enrollment.user ? enrollment.user.Email : "")
        data.push(enrollment.user ? enrollment.user.StudentID : "")
        data.push(getFormattedTime(enrollment.lastActivityDate))
        data.push(enrollment.totalApproved.toString())
        data.push(enrollment.slipDaysRemaining.toString())

        if (isPending(enrollment)) {
            data.push(
                <div>
                    <i className="badge badge-primary" style={{ cursor: "pointer" }}
                        onClick={() => { actions.updateEnrollment({ enrollment, status: Enrollment_UserStatus.STUDENT }) }}>
                        Accept
                    </i>
                    <i className="badge badge-danger clickable ml-1"
                        onClick={() => actions.updateEnrollment({ enrollment, status: Enrollment_UserStatus.NONE })}>
                        Reject
                    </i>
                </div>)
        } else {
            data.push(edit ? (
                <div>
                    <i className={`badge badge-${isTeacher(enrollment) ? "warning" : "primary"} clickable`}
                        onClick={() => actions.updateEnrollment({ enrollment, status: isTeacher(enrollment) ? Enrollment_UserStatus.STUDENT : Enrollment_UserStatus.TEACHER })}>
                        {isTeacher(enrollment) ? "Demote" : "Promote"}
                    </i>
                    <i className="badge badge-danger clickable ml-1"
                        onClick={() => actions.updateEnrollment({ enrollment, status: Enrollment_UserStatus.NONE })}>
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
                    <div className={edit ? "btn btn-danger" : "btn btn-primary"} onClick={() => setEditing(!edit)}>
                        {edit ? "Done" : "Edit"}
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
