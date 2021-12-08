import React, { useEffect } from "react"
import { User } from "../../../proto/ag/ag_pb"
import { isHidden } from "../../Helpers"
import { useAppState, useActions } from "../../overmind"
import DynamicTable, { CellElement } from "../DynamicTable"
import Search from "../Search"
import UserElement from "./User"


const Users = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        if (state.allUsers.length == 0) {
            actions.getUsers()
        }
    }, [state.query])

    const PromoteButton = ({user}: {user: User}) => {
        const classname = user.getIsadmin() ? "badge badge-danger float-right" : "badge badge-primary float-right"
        const text = user.getIsadmin() ? "Demote" : "Promote"
        return (
            <span className={classname} style={{cursor: "pointer"}} onClick={() => { actions.updateAdmin(user) }}>
                {text}
            </span>
        )
    }


    const headers: string[] = ["Name", "Email", "Student ID", "Role"]
    const users = state.allUsers.map((user, index) => {
        const data: (string | JSX.Element | CellElement)[] = []
        data.push(<UserElement user={user} hidden={isHidden(user.getName(), state.query)} />)
        data.push(user.getEmail())
        data.push(user.getStudentid())
        data.push(<PromoteButton user={user} />)
        return data
    })

    return (
        <div className="box">
            <div className="pb-2">
                <Search />
            </div>
            <DynamicTable header={headers} data={users} />
        </div>
    )
}

export default Users