import React from "react"
import { User as pbUser } from "../../../proto/qf/types_pb"


const User = ({ user }: { user: pbUser, hidden: boolean }): JSX.Element => {
    return (
        <>
            {user.Name}
            {user.IsAdmin ?
                <span className={"badge badge-primary ml-2"}>
                    Admin
                </span>
                : null
            }
        </>
    )
}

export default User
