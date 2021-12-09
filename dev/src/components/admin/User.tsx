import React from "react"
import { User } from "../../../proto/ag/ag_pb"

const UserElement = ({user, hidden}: {user: User, hidden: boolean}): JSX.Element => {
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

export default UserElement