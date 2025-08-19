import React from "react";
import { useAppState } from "../../overmind";
const ProfileCard = ({ children }) => {
    const self = useAppState().self;
    return (React.createElement("div", { className: "card", style: { width: "28rem" } },
        React.createElement("div", { className: "card-header text-center bg-dark", style: { height: "5rem", marginBottom: "3rem" } },
            React.createElement("img", { className: "card-img", style: { borderRadius: "50%", width: "8rem" }, src: self.AvatarURL, alt: "Profile Card" })),
        React.createElement("div", { className: "card-body" }, children)));
};
export default ProfileCard;
