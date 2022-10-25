import React from "react"
import { User as pbUser } from "../../../gen/qf/types_pb"


const User = ({ user }: { user: pbUser, hidden: boolean }): JSX.Element => {
    return (
        <>
            {user.name}
            {user.isAdmin ?
                <span className={"badge badge-primary ml-2"}>
                    Admin
                </span>
                : null
            }
        </>
    )
}

export default User
