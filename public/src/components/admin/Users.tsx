import React, { useEffect, useState } from "react"
import { User } from "../../../proto/ag/ag_pb"
import { useOvermind } from "../../overmind"


const Users = () => {
    const {state, actions} = useOvermind()

    const [query, setQuery] = useState<string>("")

    useEffect(() => {
        if (state.allUsers.length == 0) {
            actions.getUsers()
        }
    })

    const PromoteButton = ({user}: {user: User}) => {
        let classname = user.getIsadmin() ? "badge badge-danger float-right" : "badge badge-primary float-right"
        let text = user.getIsadmin() ? "Demote" : "Promote"
        return (
            <span className={classname} style={{cursor: "pointer"}} onClick={() => actions.updateAdmin(user)}>
                {text}
            </span>
        )
    }

    const UserListElement = ({user}: {user: User}) => {
        return (
            <li className={"list-group-item" } hidden={!user.getName().toLowerCase().includes(query.toLowerCase())}>
                {user.getName()} 
                {user.getIsadmin() ? 
                    <span className={"badge badge-primary"}>
                        Admin
                    </span> 
                    : null
                }
                <PromoteButton user={user}></PromoteButton>
            </li>
        )
    }

    const users = state.allUsers.map((user, index) => {
        return <UserListElement user={user} key={index}></UserListElement>
    })

    return (
        <div className="box">
        <ul>
        <input onKeyUp={e => setQuery(e.currentTarget.value)} placeholder={"Search"}></input>
            {users}
        </ul>
        </div>
    )
}

export default Users