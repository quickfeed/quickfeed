import React from "react"
import { Enrollment, User as pbUser } from "../../../proto/qf/types_pb"
import { useGrpc } from "../../overmind"
import { EnrollmentStatus } from "../../Helpers"
import Badge from "../Badge"

const User = ({ user }: { user: pbUser; hidden: boolean }) => {
    const { api } = useGrpc().global
    const [enrollments, setEnrollments] = React.useState<Enrollment[]>([])
    const [showEnrollments, setShowEnrollments] = React.useState<boolean>(false)

    const toggleEnrollments = () => {
        setShowEnrollments(!showEnrollments)
        if (!enrollments.length) {
            getEnrollments()
        }
    }

    const getEnrollments = () => {
        api.client
            .getEnrollments({
                FetchMode: { case: "userID", value: user.ID },
            })
            .then((response) => {
                setEnrollments(response.message.enrollments)
            })
    }

    const enrollmentsList = enrollments.length ? (
        enrollments.map((enrollment) => (
            <div key={enrollment.ID.toString()} className="mt-1">
                <Badge color="gray" text={enrollment.course?.name ?? ""} />
                <Badge className="ml-2" color={enrollment.status} text={EnrollmentStatus[enrollment.status]} />
            </div>
        ))
    ) : (
        <Badge color="gray" text="No enrollments" />
    )

    return (
        <div role="button" aria-hidden="true" className="clickable" onClick={toggleEnrollments}>
            {user.Name}
            {user.IsAdmin ? (
                <Badge color="blue" text="Admin" />
            ) : null}
            {showEnrollments ? (
                <div className="mt-1">
                    {enrollmentsList}
                </div>
            ) : null}
        </div>
    )
}

export default User
