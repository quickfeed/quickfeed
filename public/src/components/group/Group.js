import React from "react";
import { isPendingGroup } from "../../Helpers";
import { useAppState } from "../../overmind";
import { useCourseID } from "../../hooks/useCourseID";
const GroupComponent = () => {
    const state = useAppState();
    const courseID = useCourseID();
    const group = state.userGroup[courseID.toString()];
    const pendingIcon = isPendingGroup(group) ? React.createElement("span", { className: "badge badge-warning ml-2" }, "Pending") : null;
    const members = group.users.map(user => React.createElement("li", { key: user.ID.toString(), className: "list-group-item" },
        React.createElement("img", { src: user.AvatarURL, style: { width: "23px", marginRight: "10px", borderRadius: "50%" }, alt: "" }),
        user.Name));
    return (React.createElement("div", null,
        React.createElement("li", { className: "list-group-item active" },
            group.name,
            pendingIcon),
        members));
};
export default GroupComponent;
