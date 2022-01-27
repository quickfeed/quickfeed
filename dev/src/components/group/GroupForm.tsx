
import React, { useEffect, useState } from "react"
import { useHistory } from "react-router"
import { Enrollment, User } from "../../../proto/ag/ag_pb"
import { getCourseID, isApprovedGroup, isHidden, isTeacher, sortByField } from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Search from "../Search"


const GroupForm = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const history = useHistory()
    const [query, setQuery] = useState<string>("")
    const courseID = getCourseID()

    const group = state.group.group

    useEffect(() => {
        if (!state.isTeacher && !state.group.users.includes(state.self.getId())) {
            actions.group.addUsers(state.self)
        }
    }, [])

    const search = (enrollment: Enrollment): boolean => {
        if (enrollment.getUserid() in state.group.users || enrollment.hasGroup() && enrollment.getGroupid() != group?.getId()) {
            return true
        }
        if (enrollment.getUser()) {
            return isHidden((enrollment.getUser() as User).getName(), query)
        }
        return false
    }

    const sortedEnrollments = sortByField(state.courseEnrollments[courseID], [Enrollment.prototype.getUser], User.prototype.getName) as Enrollment[]

    const AvailableUser = ({ enrollment }: { enrollment: Enrollment }) => {
        const user = enrollment.getUser()
        if (!user || state.group.users.includes(user.getId()) || isTeacher(enrollment)) {
            return null
        }
        return (
            <li hidden={search(enrollment)} className="list-group-item">
                {enrollment.getUser()?.getName()}
                <i className="badge-pill badge-success ml-2 clickable float-right" onClick={() => actions.group.updateUsers(user)}>+</i>
            </li>
        )
    }

    const GroupMember = ({ enrollment }: { enrollment: Enrollment }) => {
        const user = enrollment.getUser()
        if (!user || !state.group.users.includes(user.getId())) {
            return null
        }
        return (
            <li className="list-group-item">
                <img id="group-image" src={enrollment.getUser()?.getAvatarurl()}></img>
                {enrollment.getUser()?.getName()}
                <i className="badge-pill badge-danger rounded-circle clickable float-right" onClick={() => actions.group.updateUsers(user)}>-</i>
            </li>
        )
    }

    const handleCreate = async () => {
        const result = await actions.group.createGroup()
        if (!state.isTeacher && result) {
            history.push(`/course/${courseID}/group`)
        }
    }

    const handleUpdateGroup = async () => {
        if (await actions.group.updateGroup()) {
            actions.group.resetGroup()
        }
    }

    const GroupNameBanner = <div className="card-header" style={{ textAlign: "center" }}>{state.group.name}</div>
    const GroupNameInput = group && isApprovedGroup(group)
        ? null
        : <input placeholder={"Group Name:"} onKeyUp={e => actions.group.setName(e.currentTarget.value)}></input>

    return (
        <div className="container">
            <div className="row">
                <div className="card well col-md-offset-2">
                    <div className="card-header" style={{ textAlign: "center" }}>
                        Students
                    </div>
                    <Search placeholder={"Search"} setQuery={setQuery} />

                    <ul className="list-group list-group-flush">
                        {sortedEnrollments.map((enrollment) => {
                            return <AvailableUser key={enrollment.getId()} enrollment={enrollment} />
                        })}
                    </ul>
                </div>

                <div className='col'>
                    <div className="card well col-md-offset-2" >
                        {GroupNameBanner}
                        {GroupNameInput}
                        {state.courseEnrollments[courseID].map((enrollment) => {
                            return <GroupMember key={enrollment.getId()} enrollment={enrollment} />
                        })}
                        {group.getId() > 0 ?
                            <div className="row justify-content-md-center">
                                <div className="btn btn-primary ml-2" onClick={async () => { await handleUpdateGroup() }}> Update </div>
                                <div className="btn btn-danger ml-2" onClick={() => actions.group.resetGroup()}> Cancel </div>
                            </div>
                            :
                            <div className="btn btn-primary" onClick={() => handleCreate()}> Create Group </div>
                        }
                    </div>
                </div>
            </div>
        </div>
    )
}

export default GroupForm
