import React from "react"
import Avatar from "../Avatar"
import { useAppState } from "../../overmind"


/** ProfileCard displays the profile information of the provided children as a card. */
const ProfileCard = ({ children }: { children: React.ReactNode }) => {
    const self = useAppState().self

    return (
        <div className="card bg-base-200 shadow-xl w-full max-w-md">
            <div className="card-body items-center text-center pt-8">
                <div className="mb-4">
                    <Avatar src={self.AvatarURL} alt="Profile avatar" size="w-32" />
                </div>
                {children}
            </div>
        </div>
    )
}

export default ProfileCard
