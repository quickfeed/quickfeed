import React, { useState, useCallback } from "react"
import { Color, EnrollmentSort, EnrollmentStatus, EnrollmentStatusBadge, getFormattedTime, isPending, sortEnrollments } from "../Helpers"
import { useAppState, useActions } from "../overmind"
import { Enrollment, Enrollment_UserStatus } from "../../proto/qf/types_pb"
import Search from "./Search"
import DynamicTable, { Row } from "./DynamicTable"
import DynamicButton from "./DynamicButton"
import Button, { ButtonType } from "./admin/Button"
import { useCourseID } from "../hooks/useCourseID"

const Members = () => {
    const state = useAppState()
    const actions = useActions().global
    const courseID = useCourseID()

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

    const handleMemberChange = useCallback((enrollment: Enrollment, status: Enrollment_UserStatus) => (
        () => actions.updateEnrollment({ enrollment, status })
    ), [actions])
    const handleApprovePendingEnrollments = useCallback(() => actions.approvePendingEnrollments(), [actions])

    const members = sortEnrollments(enrollments, sortBy, descending).map(enrollment => {
        // Button color and text are determined by the enrollment status
        // These are used to determine what action we can take on the enrollment
        // and what the button should say
        let buttonColor = Color.GREEN
        let enrollmentButtonText = ""
        let role = Enrollment_UserStatus.STUDENT
        switch (enrollment.status) {
            case Enrollment_UserStatus.PENDING:
                // if the enrollment is pending, we can accept them as a student
                role = Enrollment_UserStatus.STUDENT
                enrollmentButtonText = "Accept"
                buttonColor = Color.GREEN
                break
            case Enrollment_UserStatus.STUDENT:
                // if the enrollment is a student, we can promote them to teacher
                role = Enrollment_UserStatus.TEACHER
                enrollmentButtonText = "Promote"
                buttonColor = Color.BLUE
                break
            case Enrollment_UserStatus.TEACHER:
                // if the enrollment is a teacher, we can demote them to student
                role = Enrollment_UserStatus.STUDENT
                enrollmentButtonText = "Demote"
                buttonColor = Color.YELLOW
                break
            default:
                // we do not handle the case where the enrollment is NONE
                // as this status is only used by the server to reject (delete) enrollments
                // if the enrollment has any other status, we should not do anything
                role = enrollment.status
                break
        }

        const buttons = (
            <div className="d-flex">
                <DynamicButton
                    text={enrollmentButtonText}
                    color={buttonColor}
                    type={ButtonType.BADGE}
                    className="mr-2"
                    onClick={handleMemberChange(enrollment, role)}
                />
                <DynamicButton
                    text={"Reject"}
                    color={Color.RED}
                    type={ButtonType.BADGE}
                    onClick={handleMemberChange(enrollment, Enrollment_UserStatus.NONE)}
                />
            </div>
        )
        const enrollmentBadgeIcon = (
            <i className={EnrollmentStatusBadge[enrollment.status]}>
                {EnrollmentStatus[enrollment.status]}
            </i>
        )
        // rolebuttons can either be accept/reject, promote/demote or just the badge icon (student/teacher)
        const roleButtons = isPending(enrollment) || edit ? buttons : enrollmentBadgeIcon
        const { Name = "", Email = "", StudentID = "" } = enrollment.user || {}
        return [
            Name, Email, StudentID,
            getFormattedTime(enrollment.lastActivityDate),
            enrollment.totalApproved.toString(),
            enrollment.slipDaysRemaining.toString(),
            roleButtons,
        ]
    })
    return (
        <>
            <div className="row no-gutters pb-2">
                <div className="col-md-6">
                    <Search />
                </div>
                <div className="ml-auto">
                    <Button
                        text={edit ? "Done" : "Edit"}
                        color={edit ? Color.RED : Color.BLUE}
                        type={ButtonType.BUTTON}
                        onClick={() => setEditing(!edit)}
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
        </>
    )
}

export default Members
