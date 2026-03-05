import React from "react"
import { useAppState } from "../../overmind"


/** ProfileCard displays the profile information of the provided children as a card. */
const ProfileCard = ({ children }: { children: React.ReactNode }) => {
    const self = useAppState().self

    return (
        <div className="card bg-base-200 shadow-xl w-full max-w-md">
            <div className="card-body items-center text-center pt-8">
                <div className="avatar mb-4">
                    <div className="w-32 rounded-full ring ring-primary ring-offset-base-100 ring-offset-2">
                        <img src={self.AvatarURL} alt="Profile avatar" />
                    </div>
                </div>
                {children}
            </div>
        </div>
    )
}

export default ProfileCard
