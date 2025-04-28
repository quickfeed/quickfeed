import React, { useCallback, useEffect } from "react"
import { isHidden, Color, userLink } from "../../Helpers"
import { useAppState, useActions } from "../../overmind"
import DynamicButton from "../DynamicButton"
import DynamicTable from "../DynamicTable"
import Search from "../Search"
import { ButtonType } from "./Button"
import UserComponent from "./User"
import { User } from "../../../proto/qf/types_pb"


const Users = () => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        actions.getUsers()
    }, [])

    const handlePromoteAdmin = useCallback((user: User) => () => actions.updateAdmin(user), [actions])

    const headers: string[] = ["Name", "GitHub", "Email", "Student ID", "Role"]
    const users = state.allUsers.map((user) => {
        const roleActionText = user.IsAdmin ? "Demote" : "Promote"
        const buttonColor = user.IsAdmin ? Color.RED : Color.BLUE
        return [
            <UserComponent key={user.ID} user={user} hidden={!isHidden(user.Name, state.query)} />,
            <a key={`${user.ID}link`} href={userLink(user)}>{user.Login}</a>,
            user.Email,
            user.StudentID,
            <DynamicButton
                key={`${user.ID}button`}
                text={roleActionText}
                color={buttonColor}
                type={ButtonType.BADGE}
                onClick={handlePromoteAdmin(user)}
            />
        ]
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
