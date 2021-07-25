import React, { useEffect, useState } from "react"
import { User } from "../../../proto/ag/ag_pb"
import { useOvermind } from "../../overmind"
import Search from "../Search"


const Users = () => {
    const {state, actions} = useOvermind()

    useEffect(() => {
        if (state.allUsers.length == 0) {
            actions.getUsers()
        }
    }, [state.query])

    const PromoteButton = (props: {user: User, onClick?: Function, input?: any}) => {
        const classname = props.user.getIsadmin() ? "badge badge-danger float-right" : "badge badge-primary float-right"
        const text = props.user.getIsadmin() ? "Demote" : "Promote"
        return (
            <span className={classname} style={{cursor: "pointer"}} onClick={() => {if (props.onClick) { props.onClick(props.input)} }}>
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
                <PromoteButton user={user} onClick={actions.updateAdmin} input={user}></PromoteButton>
            </li>
        )
    }

    const users = state.allUsers.map((user, index) => {
        return <UserListElement user={user} key={index}></UserListElement>
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