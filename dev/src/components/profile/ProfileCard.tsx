import React from "react"
import { useAppState } from "../../overmind"

/** ProfileCard displays the profile information of the provided children as a card. */
const ProfileCard = ({ children }: { children: React.ReactNode }): JSX.Element => {
    const self = useAppState().self

    return (
        <div className="card" style={{ width: "28rem" }}>
            <div className="card-header text-center bg-dark" style={{ height: "5rem", marginBottom: "3rem" }}>
                <img className="card-img" style={{ borderRadius: "50%", width: "8rem" }} src={self.getAvatarurl()} alt="Profile Image"></img>
            </div>
            <div className="card-body">
                {children}
            </div>
        </div>
    )
}

export default ProfileCard
