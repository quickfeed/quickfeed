import { clone, create } from "@bufbuild/protobuf"
import React, { useEffect, useState } from "react"
import { Enrollment, Enrollment_UserStatus, EnrollmentSchema, GroupSchema, UserSchema } from "../../../proto/qf/types_pb"
import { Color, hasTeacher, isApprovedGroup, isHidden, isPending, isStudent } from "../../Helpers"
import { useCourseID } from "../../hooks/useCourseID"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import DynamicButton from "../DynamicButton"
import Search from "../Search"


const GroupForm = () => {
    const state = useAppState()
    const actions = useActions().global

    const [query, setQuery] = useState<string>("")
    const [enrollmentType, setEnrollmentType] = useState<Enrollment_UserStatus.STUDENT | Enrollment_UserStatus.TEACHER>(Enrollment_UserStatus.STUDENT)
    const courseID = useCourseID()

    const group = state.activeGroup
    useEffect(() => {
        if (isStudent(state.enrollmentsByCourseID[courseID.toString()])) {
            actions.setActiveGroup(create(GroupSchema))
            actions.updateGroupUsers(clone(UserSchema, state.self))
        }
        return () => {
            actions.setActiveGroup(null)
        }
    }, [actions, courseID, state.enrollmentsByCourseID, state.self])
    if (!group) {
        return null
    }
    const userIds = group.users.map(user => user.ID)

    const search = (enrollment: Enrollment): boolean => {
        if (userIds.includes(enrollment.userID) || enrollment.group && enrollment.groupID !== group.ID) {
            return true
        }
        if (enrollment.user) {
            return isHidden(enrollment.user.Name, query)
        }
        return false
    }

    const enrollments = state.courseEnrollments[courseID.toString()].map(enrollment => clone(EnrollmentSchema, enrollment))

    // Determine the user's enrollment status (teacher or student)
    const isTeacher = hasTeacher(state.status[courseID.toString()])

    const enrollmentFilter = (enrollment: Enrollment) => {
        if (isTeacher) {
            // If the user is a teacher, show all enrollments of the selected enrollment type
            return enrollment.status === enrollmentType
        }
        // Show all students
        return enrollment.status === Enrollment_UserStatus.STUDENT
    }

    const groupFilter = (enrollment: Enrollment) => {
        if (group && group.ID) {
            // If a group is being edited, show users that are in the group
            // This is to allow users to be removed from the group, and to be re-added
            return enrollment.groupID === group.ID || enrollment.groupID === BigInt(0)
        }
        // Otherwise, show users that are not in a group
        return enrollment.groupID === BigInt(0)
    }

    const sortedAndFilteredEnrollments = enrollments
        // Filter enrollments where the user is not a student (or teacher), or the user is already in a group
        .filter(enrollment => enrollmentFilter(enrollment) && groupFilter(enrollment))
        // Sort by name
        .sort((a, b) => (a.user?.Name ?? "").localeCompare((b.user?.Name ?? "")))

    const AvailableUser = ({ enrollment }: { enrollment: Enrollment }) => {
        const id = enrollment.userID
        if (isPending(enrollment)) {
            return null
        }
        if (id !== state.self.ID && !userIds.includes(id)) {
            return (
                <div
                    hidden={search(enrollment)}
                    key={id.toString()}
                    className="flex items-center justify-between p-3 hover:bg-base-200 rounded-lg transition-colors group"
                >
                    <span className="font-medium">{enrollment.user?.Name}</span>
                    <button
                        className="btn btn-sm btn-circle btn-success opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={() => actions.updateGroupUsers(enrollment.user)}
                    >
                        <i className="fa fa-plus"></i>
                    </button>
                </div>
            )
        }
        return null
    }

    const groupMembers = group.users.map(user => {
        return (
            <div
                key={user.ID.toString()}
                className="flex items-center justify-between p-3 bg-base-200/50 hover:bg-base-200 rounded-lg transition-colors group"
            >
                <div className="flex items-center gap-3">
                    <div className="avatar">
                        <div className="w-10 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2">
                            <img src={user.AvatarURL} alt={`${user.Name}'s avatar`} />
                        </div>
                    </div>
                    <span className="font-medium">{user.Name}</span>
                </div>
                <button
                    className="btn btn-sm btn-circle btn-error opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={() => actions.updateGroupUsers(user)}
                >
                    <i className="fa fa-times"></i>
                </button>
            </div>
        )
    })

    const toggleEnrollmentType = () => {
        if (hasTeacher(enrollmentType)) {
            setEnrollmentType(Enrollment_UserStatus.STUDENT)
        } else {
            setEnrollmentType(Enrollment_UserStatus.TEACHER)
        }
    }

    const EnrollmentTypeButton = () => {
        if (!isTeacher) {
            return (
                <div className="flex items-center justify-center gap-2 text-lg font-semibold">
                    <i className="fa fa-users"></i>
                    <span>Students</span>
                </div>
            )
        }
        return (
            <button className="btn btn-primary w-full gap-2" type="button" onClick={toggleEnrollmentType}>
                <i className={`fa ${enrollmentType === Enrollment_UserStatus.STUDENT ? 'fa-user-graduate' : 'fa-chalkboard-teacher'}`}></i>
                {enrollmentType === Enrollment_UserStatus.STUDENT ? "Students" : "Teachers"}
            </button>
        )
    }

    const GroupNameBanner = (
        <div className="flex items-center justify-center gap-2 bg-primary text-primary-content px-4 py-3 rounded-t-2xl">
            <i className="fa fa-users"></i>
            <h3 className="text-lg font-bold">{group.name}</h3>
        </div>
    )

    const GroupNameInput = group && isApprovedGroup(group)
        ? null
        : (
            <div className="form-control px-4 pt-4">
                <input
                    className="input input-bordered w-full focus:input-primary"
                    placeholder="Enter group name..."
                    onKeyUp={e => actions.updateGroupName(e.currentTarget.value)}
                />
            </div>
        )

    return (
        <div className="container mx-auto px-4 py-8">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-6xl mx-auto">
                {/* Available Users Panel */}
                <div className="card bg-base-200 shadow-xl">
                    <div className="card-body p-0">
                        <div className="bg-base-200 px-4 py-3 rounded-t-2xl">
                            <EnrollmentTypeButton />
                        </div>
                        <div className="px-4 pt-4">
                            <Search placeholder="Search users..." setQuery={setQuery} />
                        </div>
                        <div className="p-4 space-y-1 max-h-96 overflow-y-auto">
                            {sortedAndFilteredEnrollments.length > 0 ? (
                                sortedAndFilteredEnrollments.map((enrollment) => {
                                    return <AvailableUser key={enrollment.ID} enrollment={enrollment} />
                                })
                            ) : (
                                <div className="text-center py-8 text-base-content/60">
                                    <i className="fa fa-users text-3xl mb-2"></i>
                                    <p>No users available</p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Group Members Panel */}
                <div className="card bg-base-200 shadow-xl">
                    <div className="card-body p-0">
                        {GroupNameBanner}
                        {GroupNameInput}
                        <div className="p-4 space-y-2">
                            {groupMembers.length > 0 ? (
                                groupMembers
                            ) : (
                                <div className="text-center py-8 text-base-content/60">
                                    <i className="fa fa-user-plus text-3xl mb-2"></i>
                                    <p>Add members to your group</p>
                                </div>
                            )}
                        </div>
                        <div className="px-4 pb-4 pt-2 border-t border-base-300">
                            {group && group.ID ? (
                                <div className="flex gap-3">
                                    <DynamicButton
                                        text="Update Group"
                                        color={Color.BLUE}
                                        className="flex-1"
                                        onClick={() => actions.updateGroup(group)}
                                    />
                                    <Button
                                        text="Cancel"
                                        color={Color.RED}
                                        type={ButtonType.OUTLINE}
                                        onClick={() => actions.setActiveGroup(null)}
                                    />
                                </div>
                            ) : (
                                <DynamicButton
                                    text="Create Group"
                                    color={Color.GREEN}
                                    className="w-full"
                                    onClick={() => actions.createGroup({ courseID, users: userIds, name: group.name })}
                                />
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GroupForm
