import React from "react"
import { Enrollment, User as pbUser } from "../../../proto/qf/types_pb"
import { useGrpc } from "../../overmind"
import { EnrollmentStatus, EnrollmentStatusBadgeColor } from "../../Helpers"

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
        <div className="grid gap-2 mt-3">
            {enrollments.map((enrollment) => (
                <div
                    key={enrollment.ID.toString()}
                    className="flex items-center justify-between p-3 bg-base-300 rounded-lg hover:bg-base-200 transition-colors"
                >
                    <div className="flex items-center gap-3">
                        <div>
                            <div className="font-medium">{enrollment.course?.name ?? "Unknown Course"}</div>
                            <div className="text-xs text-base-content/60">
                                {enrollment.course?.code} · {enrollment.course?.year}
                            </div>
                        </div>
                    </div>
                    <span className={`badge ${EnrollmentStatusBadgeColor[enrollment.status]}`}>
                        {EnrollmentStatus[enrollment.status]}
                    </span>
                </div>
            ))}
        </div>
    ) : (
        <div className="text-base-content/60 text-sm mt-2">No enrollments</div>
    )

    return (
        <div
            role="button"
            tabIndex={0}
            className="p-3 rounded-lg hover:bg-base-200 transition-colors cursor-pointer"
            onClick={toggleEnrollments}
            onKeyDown={(e) => e.key === 'Enter' && toggleEnrollments()}
        >
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <div className="avatar">
                        <div className="w-10 rounded-full ring ring-base-300">
                            <img src={user.AvatarURL} alt={user.Name} />
                        </div>
                    </div>
                    <div>
                        <div className="font-semibold">{user.Name}</div>
                        <div className="text-sm text-base-content/60">{user.Login}</div>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    {user.IsAdmin && (
                        <span className="badge badge-primary">Admin</span>
                    )}
                    <i className={`fa fa-chevron-down transition-transform ${showEnrollments ? 'rotate-180' : ''}`} />
                </div>
            </div>
            {showEnrollments && enrollmentsList}
        </div>
    )
}

export default User
