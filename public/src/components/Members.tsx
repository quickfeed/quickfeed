import React, { useCallback, useState } from "react"
import { Color, EnrollmentSort, EnrollmentStatus, EnrollmentStatusBadge, getCourseID, getFormattedTime, isPending, isTeacher, sortEnrollments } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment, Enrollment_UserStatus } from "../../proto/qf/types_pb"
import Search from "./Search"
import DynamicTable, { Row } from "./DynamicTable"
import DynamicButton from "./DynamicButton"
import Button, { ButtonType } from "./admin/Button"

const Members = () => {
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

    const handleMemberChange = useCallback((enrollment: Enrollment, status: Enrollment_UserStatus) => () => actions.updateEnrollment({ enrollment, status: status }), [actions])
    const handleApprovePendingEnrollments = useCallback(() => actions.approvePendingEnrollments(), [actions])

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
                <div className="d-flex">
                    <DynamicButton
                        text={"Accept"}
                        color={Color.GREEN}
                        type={ButtonType.BADGE}
                        className="mr-2"
                        onClick={handleMemberChange(enrollment, Enrollment_UserStatus.STUDENT)}
                    />
                    <DynamicButton
                        text={"Reject"}
                        color={Color.RED}
                        type={ButtonType.BADGE}
                        onClick={handleMemberChange(enrollment, Enrollment_UserStatus.NONE)}
                    />
                </div>)
        } else {
            data.push(edit ? (
                <div className="d-flex">
                    <DynamicButton
                        text={isTeacher(enrollment) ? "Demote" : "Promote"}
                        color={isTeacher(enrollment) ? Color.YELLOW : Color.BLUE}
                        type={ButtonType.BADGE}
                        className="mr-2"
                        onClick={handleMemberChange(enrollment, isTeacher(enrollment) ? Enrollment_UserStatus.STUDENT : Enrollment_UserStatus.TEACHER)}
                    />
                    <DynamicButton
                        text={"Reject"}
                        color={Color.RED}
                        type={ButtonType.BADGE}
                        onClick={handleMemberChange(enrollment, Enrollment_UserStatus.NONE)}
                    />
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
                    <Button
                        text={edit ? "Done" : "Edit"}
                        color={edit ? Color.RED : Color.BLUE}
                        type={ButtonType.BUTTON}
                        onClick={() => setEditing(!edit)} // skipcq: JS-0417
                    />
                </div>
                {pending?.length > 0 ?
                    <div style={{ marginLeft: "10px" }}>
                        <DynamicButton
                            text="Approve All"
                            color={Color.GREEN}
                            type={ButtonType.BUTTON}
                            onClick={handleApprovePendingEnrollments}
                        />
                    </div> : null}
            </div>

            <div>
                <DynamicTable header={header} data={members} />
            </div>
        </div>
    )
}

export default Members
