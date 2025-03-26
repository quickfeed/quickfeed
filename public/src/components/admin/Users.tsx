import React, { useEffect, useCallback } from "react"
import { isHidden, Color, userLink } from "../../Helpers"
import { useAppState, useActions } from "../../overmind"
import DynamicButton from "../DynamicButton"
import DynamicTable, { Row } from "../DynamicTable"
import Search from "../Search"
import { ButtonType } from "./Button"
import UserComponent from "./User"
import { User } from "../../../proto/qf/types_pb"


const Users = () => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        actions.getUsers()
    }, [actions])

    const handleUpdateAdmin = useCallback((user: User) => () => actions.updateAdmin(user), [actions])

    const headers: string[] = ["Name", "GitHub", "Email", "Student ID", "Role"]
    const users = state.allUsers.map((user) => {
        const data: Row = []
        data.push(<UserComponent user={user} hidden={!isHidden(user.Name, state.query)} />)
        data.push(<a href={userLink(user)}>{user.Login}</a>)
        data.push(user.Email)
        data.push(user.StudentID)
        data.push(
            <DynamicButton
                text={user.IsAdmin ? "Demote" : "Promote"}
                color={user.IsAdmin ? Color.RED : Color.BLUE}
                type={ButtonType.BADGE}
                onClick={handleUpdateAdmin(user)}
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
