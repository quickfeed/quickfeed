import React from "react"
import { User } from "../../../proto/ag/ag_pb"

const UserElement = ({user, hidden}: {user: User, hidden: boolean}): JSX.Element => {
    return (
        <div>
            {user.getName()} 
            {user.getIsadmin() ? 
                <span className={"badge badge-primary ml-2"}>
                    Admin
                </span> 
                : null
            }
        </div>
    )
}

export default UserElement