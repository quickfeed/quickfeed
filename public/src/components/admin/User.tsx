import React from "react"
import { User as user } from "../../../proto/ag/ag_pb"


const User = ({ user, hidden }: { user: user.AsObject, hidden: boolean }): JSX.Element => {
    return (
        <>
            {user.name}
            {user.isadmin ?
                <span className={"badge badge-primary ml-2"}>
                    Admin
                </span>
                : null
            }
        </>
    )
}

export default User
