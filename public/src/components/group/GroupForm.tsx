import { clone, create } from "@bufbuild/protobuf"
import type { KeyboardEvent } from "react"
import { useEffect, useState } from "react"
import type { Enrollment, Group, User } from "../../../proto/qf/types_pb"
import { Enrollment_UserStatus, GroupSchema, UserSchema } from "../../../proto/qf/types_pb"
import { Color, groupNameError, hasTeacher, isApprovedGroup, isPendingGroup, isStudent, maxGroupNameLength } from "../../Helpers"
import { useCourseID } from "../../hooks/useCourseID"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import Avatar from "../Avatar"
import DynamicButton from "../DynamicButton"


type Role = Enrollment_UserStatus.STUDENT | Enrollment_UserStatus.TEACHER

/** Candidate is a course member that can be listed in the member picker. */
interface Candidate {
    user: User
    selected: boolean
    /** Name of another group the user belongs to, or "" if the user is free to join. */
    otherGroup: string
    /** Locked members cannot be removed, e.g., a student creating their own group. */
    locked: boolean
}

const GroupForm = () => {
    const state = useAppState()
    const actions = useActions().global
    const courseID = useCourseID()

    const [query, setQuery] = useState<string>("")
    const [role, setRole] = useState<Role>(Enrollment_UserStatus.STUDENT)

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

    const isTeacher = hasTeacher(state.status[courseID.toString()])
    const isEditing = group.ID > 0n
    const memberIDs = group.users.map(user => user.ID)
    // Only members may create a group, so a student cannot leave the group they are creating.
    const isLocked = (userID: bigint) => !isTeacher && userID === state.self.ID

    // A teacher's group list stays current after groups are edited or deleted, unlike enrollments' group IDs.
    // Students do not receive the group list, so for them the enrollments are all there is.
    const courseGroups = isTeacher ? state.groups[courseID.toString()] : undefined
    const groupByUser = new Map<bigint, Group>()
    for (const g of courseGroups ?? []) {
        for (const user of g.users) {
            groupByUser.set(user.ID, g)
        }
    }
    const otherGroupName = (enrollment: Enrollment): string => {
        if (courseGroups) {
            const g = groupByUser.get(enrollment.userID)
            return g && g.ID !== group.ID ? g.name : ""
        }
        if (enrollment.groupID === 0n || enrollment.groupID === group.ID) {
            return ""
        }
        return enrollment.group?.name ?? "another group"
    }

    const enrollments = state.courseEnrollments[courseID.toString()] ?? []
    const toCandidate = (enrollment: Enrollment): Candidate | null => {
        if (!enrollment.user) {
            return null
        }
        return {
            user: enrollment.user,
            selected: memberIDs.includes(enrollment.userID),
            otherGroup: otherGroupName(enrollment),
            locked: isLocked(enrollment.userID),
        }
    }
    const countByRole = (r: Role) => enrollments.filter(e => e.status === r).length
    const candidates = enrollments
        .filter(enrollment => enrollment.status === role)
        .map(toCandidate)
        .filter((c): c is Candidate => c !== null)
        .filter(c => matches(c.user, query))
        // Users who can join come first; users in other groups sink to the bottom.
        .sort((a, b) => Number(a.otherGroup !== "") - Number(b.otherGroup !== "") || a.user.Name.localeCompare(b.user.Name))

    const toggle = (user: User) => actions.updateGroupUsers(clone(UserSchema, user))

    // Enter adds the only person left in the search results, so a member can be added without the mouse.
    const handleSearchKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key !== "Enter") {
            return
        }
        const addable = candidates.filter(c => !c.selected && !c.otherGroup)
        if (addable.length === 1) {
            toggle(addable[0].user)
            setQuery("")
        }
    }

    const nameError = group.name.length > 0 ? groupNameError(group.name) : ""
    const blocker = submitBlocker(group.name, nameError, memberIDs.length)
    const submit = () => isEditing
        ? actions.updateGroup(group)
        : actions.createGroup({ courseID, users: memberIDs, name: group.name })

    return (
        <div className="card bg-base-100 border border-base-300 shadow-sm max-w-3xl mx-auto my-6">
            <div className="card-body gap-6">
                <Header group={group} isTeacher={isTeacher} isEditing={isEditing} />

                <NameField
                    name={group.name}
                    error={nameError}
                    locked={isApprovedGroup(group)}
                    onChange={name => actions.updateGroupName(name)}
                />

                <section>
                    <h3 className="text-sm font-semibold mb-2">
                        Members <span className="text-base-content/60 font-normal">({group.users.length})</span>
                    </h3>
                    <MemberChips users={group.users} isLocked={isLocked} onRemove={toggle} />
                </section>

                <section>
                    <div className="flex flex-col sm:flex-row sm:items-center gap-2 mb-2">
                        <h3 className="text-sm font-semibold sm:mr-auto">Add members</h3>
                        {isTeacher && (
                            <RoleTabs
                                role={role}
                                setRole={setRole}
                                students={countByRole(Enrollment_UserStatus.STUDENT)}
                                teachers={countByRole(Enrollment_UserStatus.TEACHER)}
                            />
                        )}
                    </div>
                    <label className="input w-full mb-2">
                        <i className="fas fa-magnifying-glass text-base-content/50" />
                        <input
                            type="search"
                            // Students do not receive other users' GitHub usernames.
                            placeholder={isTeacher ? "Search by name or GitHub username" : "Search by name"}
                            value={query}
                            onChange={e => setQuery(e.target.value)}
                            onKeyDown={handleSearchKeyDown}
                        />
                    </label>
                    <CandidateList candidates={candidates} query={query} role={role} onToggle={toggle} />
                </section>
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center gap-3 border-t border-base-300 px-8 py-4">
                <p className="text-sm text-base-content/60 sm:mr-auto">
                    {blocker || footerNote(isTeacher, isEditing)}
                </p>
                <div className="flex gap-2 justify-end">
                    {isTeacher && (
                        <Button
                            text="Cancel"
                            color={Color.WHITE}
                            type={ButtonType.GHOST}
                            onClick={() => actions.setActiveGroup(null)}
                        />
                    )}
                    <DynamicButton
                        text={isEditing ? "Save changes" : "Create group"}
                        color={isEditing ? Color.BLUE : Color.GREEN}
                        disabled={blocker !== ""}
                        onClick={submit}
                    />
                </div>
            </div>
        </div>
    )
}

const matches = (user: User, query: string) => {
    const q = query.trim().toLowerCase()
    return q === "" || user.Name.toLowerCase().includes(q) || user.Login.toLowerCase().includes(q)
}

/** submitBlocker returns why the group cannot be submitted yet, or "" if it can. */
const submitBlocker = (name: string, nameError: string, members: number) => {
    if (name.length === 0) {
        return "Enter a group name to continue."
    }
    if (nameError) {
        return "Fix the group name to continue."
    }
    if (members === 0) {
        return "Add at least one member to continue."
    }
    return ""
}

const footerNote = (isTeacher: boolean, isEditing: boolean) => {
    if (isEditing) {
        return ""
    }
    return isTeacher
        ? "The group starts as pending. Approve it in the group list to create its repository."
        : "A teacher must approve the group before its repository is created."
}

const Header = ({ group, isTeacher, isEditing }: { group: Group, isTeacher: boolean, isEditing: boolean }) => {
    let title = "Create your group"
    let subtitle = "Choose a name and add the classmates you will work with."
    if (isEditing) {
        title = `Edit ${group.name}`
        subtitle = "Change the group's members or name."
    } else if (isTeacher) {
        title = "New group"
        subtitle = "Choose a name and pick the members."
    }
    return (
        <header>
            <div className="flex items-center gap-2">
                <h2 className="card-title text-2xl">{title}</h2>
                {isEditing && isPendingGroup(group) && <span className="badge badge-warning badge-sm">Pending</span>}
            </div>
            <p className="text-base-content/60">{subtitle}</p>
        </header>
    )
}

const NameField = ({ name, error, locked, onChange }: { name: string, error: string, locked: boolean, onChange: (name: string) => void }) => {
    let hint = "Letters, numbers, dashes and underscores. Spaces become underscores."
    if (locked) {
        hint = "The name cannot change after approval, since the group's repository is named after it."
    }
    return (
        <fieldset className="fieldset p-0">
            <legend className="fieldset-legend text-sm">Group name</legend>
            <div className="relative">
                <input
                    className={`input w-full pr-16 ${error ? "input-error" : ""}`}
                    placeholder="e.g. team_rocket"
                    value={name}
                    disabled={locked}
                    maxLength={maxGroupNameLength}
                    autoFocus={!locked}
                    // Spaces are the most common invalid character; replace them rather than reject them.
                    onChange={e => onChange(e.target.value.replace(/\s/g, "_"))}
                />
                {!locked && (
                    <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-base-content/50 tabular-nums">
                        {name.length}/{maxGroupNameLength}
                    </span>
                )}
            </div>
            <p className={`label text-xs ${error ? "text-error" : ""}`}>{error || hint}</p>
        </fieldset>
    )
}

const MemberChips = ({ users, isLocked, onRemove }: { users: User[], isLocked: (id: bigint) => boolean, onRemove: (user: User) => void }) => {
    if (users.length === 0) {
        return (
            <div className="rounded-box border border-dashed border-base-300 px-4 py-3 text-sm text-base-content/60">
                No members yet. Pick people from the list below.
            </div>
        )
    }
    return (
        <ul className="flex flex-wrap gap-2">
            {users.map(user => (
                <li key={user.ID.toString()} className="flex items-center gap-2 rounded-full bg-base-200 py-1 pl-1 pr-2">
                    <Avatar src={user.AvatarURL} alt="" size="w-7" variant="inline" />
                    <span className="text-sm font-medium">{user.Name}</span>
                    {isLocked(user.ID) ? (
                        <span className="text-xs text-base-content/60 pr-1">you</span>
                    ) : (
                        <button
                            type="button"
                            className="btn btn-ghost btn-xs btn-circle"
                            aria-label={`Remove ${user.Name}`}
                            title={`Remove ${user.Name}`}
                            onClick={() => onRemove(user)}
                        >
                            <i className="fas fa-xmark" />
                        </button>
                    )}
                </li>
            ))}
        </ul>
    )
}

const RoleTabs = ({ role, setRole, students, teachers }: { role: Role, setRole: (role: Role) => void, students: number, teachers: number }) => {
    const tab = (r: Role, label: string, count: number) => (
        <button
            type="button"
            role="tab"
            aria-selected={role === r}
            className={`tab gap-2 ${role === r ? "tab-active" : ""}`}
            onClick={() => setRole(r)}
        >
            {label}
            <span className="badge badge-sm badge-ghost">{count}</span>
        </button>
    )
    return (
        <div role="tablist" className="tabs tabs-box tabs-sm">
            {tab(Enrollment_UserStatus.STUDENT, "Students", students)}
            {tab(Enrollment_UserStatus.TEACHER, "Teachers", teachers)}
        </div>
    )
}

const CandidateList = ({ candidates, query, role, onToggle }: { candidates: Candidate[], query: string, role: Role, onToggle: (user: User) => void }) => {
    if (candidates.length === 0) {
        const who = role === Enrollment_UserStatus.TEACHER ? "teachers" : "students"
        return (
            <div className="rounded-box border border-base-300 py-8 text-center text-sm text-base-content/60">
                {query ? `No ${who} match "${query}".` : `No ${who} in this course.`}
            </div>
        )
    }
    return (
        <ul className="rounded-box border border-base-300 max-h-80 overflow-y-auto divide-y divide-base-200">
            {candidates.map(c => {
                const disabled = c.locked || c.otherGroup !== ""
                return (
                    <li key={c.user.ID.toString()}>
                        <label
                            className={`flex items-center gap-3 px-3 py-2 ${disabled ? "cursor-not-allowed" : "cursor-pointer hover:bg-base-200"}`}
                            title={c.otherGroup ? "Remove them from their current group first" : undefined}
                        >
                            <input
                                type="checkbox"
                                className="checkbox checkbox-primary checkbox-sm"
                                checked={c.selected}
                                disabled={disabled}
                                onChange={() => onToggle(c.user)}
                            />
                            <Avatar src={c.user.AvatarURL} alt="" size="w-8" variant="inline" />
                            <div className={`min-w-0 flex-1 ${c.otherGroup ? "opacity-50" : ""}`}>
                                <div className="truncate text-sm font-medium">{c.user.Name}</div>
                                {c.user.Login && <div className="truncate text-xs text-base-content/60">{c.user.Login}</div>}
                            </div>
                            {c.otherGroup && <span className="badge badge-ghost badge-sm whitespace-nowrap">In {c.otherGroup}</span>}
                        </label>
                    </li>
                )
            })}
        </ul>
    )
}

export default GroupForm
