import React from "react"
import { isHidden, Color } from "../../Helpers"
import { useAppState, useActions } from "../../overmind"
import DynamicTable, { CellElement } from "../DynamicTable"
import Search from "../Search"
import Button, { ButtonType } from "./Button"
import User from "./User"


const Users = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    const headers: string[] = ["Name", "Email", "Student ID", "Role"]
    const users = state.allUsers.map((user) => {
        const data: (string | JSX.Element | CellElement)[] = []
        data.push(
            <User user={user} hidden={!isHidden(user.getName(), state.query)} />)
        data.push(user.getEmail())
        data.push(user.getStudentid())
        data.push(
            <Button
                color={user.getIsadmin() ? Color.RED : Color.BLUE}
                type={ButtonType.BADGE}
                text={user.getIsadmin() ? "Demote" : "Promote"}
                onclick={() => actions.updateAdmin(user)}
            />
        )
        return data
    })

    return (
        <div>
            <div className="pb-2">
                <Search />
            </div>
            <DynamicTable header={headers} data={users} />
        </div>
    )
}

export default Users
