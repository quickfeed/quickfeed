import React, { useState, useCallback } from "react"
import { Enrollment, User as pbUser } from "../../../proto/qf/types_pb"
import { useGrpc } from "../../overmind"
import { EnrollmentStatus, EnrollmentStatusBadge } from "../../Helpers"

const User = ({ user }: { user: pbUser; hidden: boolean }) => {
    const { api } = useGrpc()
    const [enrollments, setEnrollments] = useState<Enrollment[]>([])
    const [showEnrollments, setShowEnrollments] = useState<boolean>(false)

    const toggleEnrollments = useCallback(() => {
        setShowEnrollments(!showEnrollments)
        if (!enrollments.length) {
            api.client
                .getEnrollments({
                    FetchMode: { case: "userID", value: user.ID },
                })
                .then((response) => {
                    setEnrollments(response.message.enrollments)
                })
        }
    }, [api.client, enrollments.length, showEnrollments, user.ID])


    const enrollmentsList = enrollments.length ? (
        <div>
            {enrollments.map((enrollment) => (
                <div key={enrollment.ID.toString()}>
                    <span className="badge badge-secondary">
                        {enrollment.course?.name}
                    </span>{" "}
                    <span className={EnrollmentStatusBadge[enrollment.status]}>
                        {EnrollmentStatus[enrollment.status]}
                    </span>
                </div>
            ))}
        </div>
    ) : (
        <div>
            <span className="badge badge-secondary">No enrollments</span>
        </div>
    )

    return (
        <div role="button" aria-hidden="true" className="clickable" onClick={toggleEnrollments}>
            {user.Name}
            {user.IsAdmin ? (
                <span className={"badge badge-primary ml-2"}>Admin</span>
            ) : null}
            {showEnrollments ? enrollmentsList : null}
        </div>
    )
}

export default User
