import React, { useEffect } from "react"
import { User } from "../../../proto/ag/ag_pb"
import { useAppState, useActions } from "../../overmind"
import Search from "../Search"


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

    const UserListElement = ({user}: {user: User}) => {
        return (
            <li className={"list-group-item" } hidden={!user.getName().toLowerCase().includes(state.query)}>
                {user.getName()} 
                {user.getIsadmin() ? 
                    <span className={"badge badge-primary"}>
                        Admin
                    </span> 
                    : null
                }
                <PromoteButton user={user} />
            </li>
        )
    }

    const users = state.allUsers.map((user, index) => {
        return <UserListElement user={user} key={index} />
    })

    return (
        <div className="box">
        <ul>
        <Search />
            {users}
        </ul>
        </div>
    )
}

export default Users