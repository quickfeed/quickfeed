import React from "react"
import { User as user } from "../../../proto/ag/ag_pb"

const User = ({ user, hidden }: { user: user, hidden: boolean }): JSX.Element => {
    return (
        <>
            {user.getName()}
            {user.getIsadmin() ?
                <span className={"badge badge-primary ml-2"}>
                    Admin
                </span>
                : null
            }
        </>
    )
}

export default User
