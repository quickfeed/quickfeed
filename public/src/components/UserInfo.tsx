import React from "react"
import { User } from "../../proto/qf/types_pb"
import Avatar from "./Avatar"

interface UserInfoProps {
    user: User
    /** Show the avatar. Defaults to true. */
    avatar?: boolean
    /** Show the login handle below the name. Defaults to false. */
    login?: boolean
}

const UserInfo = ({ user, avatar = true, login = false }: UserInfoProps) => (
    <div className="flex items-center gap-3">
        {avatar && <Avatar src={user.AvatarURL} alt={`${user.Name}'s avatar`} />}
        <div>
            <div className="font-semibold">{user.Name}</div>
            {login && <div className="text-sm text-base-content/60">{user.Login}</div>}
        </div>
    </div>
)

export default UserInfo
