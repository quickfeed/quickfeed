import React, { useEffect } from "react"
import { isHidden, Color, userLink } from "../../Helpers"
import { useAppState, useActions } from "../../overmind"
import DynamicButton from "../DynamicButton"
import DynamicTable from "../DynamicTable"
import Search from "../Search"
import { ButtonType } from "./Button"
import User from "./User"


const Users = () => {
    const state = useAppState()
    const actions = useActions()

    useEffect(() => {
        actions.getUsers()
    }, [])

    const headers: string[] = ["Name", "GitHub", "Email", "Student ID", "Role"]
    const users = state.allUsers.map((user) => {
        const roleActionText = user.IsAdmin ? "Demote" : "Promote"
        const buttonColor = user.IsAdmin ? Color.RED : Color.BLUE
        return [
            <User key={user.ID} user={user} hidden={!isHidden(user.Name, state.query)} />,
            <a key={`${user.ID}link`} href={userLink(user)}>{user.Login}</a>,
            user.Email,
            user.StudentID,
            <DynamicButton
                key={`${user.ID}button`}
                text={roleActionText}
                color={buttonColor}
                type={ButtonType.BADGE}
                onClick={() => actions.updateAdmin(user)}
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
