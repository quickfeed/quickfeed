import React, { useCallback } from "react";
import { Group_GroupStatus } from "../../proto/qf/types_pb";
import { Color, hasUsers, isApprovedGroup, isPendingGroup } from "../Helpers";
import { useActions, useAppState } from "../overmind";
import Button, { ButtonType } from "./admin/Button";
import DynamicButton from "./DynamicButton";
import GroupForm from "./group/GroupForm";
import Search from "./Search";
import { useCourseID } from "../hooks/useCourseID";
const Groups = () => {
    const state = useAppState();
    const actions = useActions().global;
    const courseID = useCourseID();
    const groupSearch = (group) => {
        if (state.query.length === 0) {
            return false;
        }
        if (group.name.toLowerCase().includes(state.query)) {
            return false;
        }
        for (const user of group.users) {
            if (user.Name.toLowerCase().includes(state.query)) {
                return false;
            }
        }
        return true;
    };
    const approveGroup = useCallback((group) => () => actions.updateGroupStatus({ group, status: Group_GroupStatus.APPROVED }), [actions]);
    const handleEditGroup = useCallback((group) => () => actions.setActiveGroup(group), [actions]);
    const handleDeleteGroup = useCallback((group) => () => actions.deleteGroup(group), [actions]);
    const GroupButtons = ({ group }) => {
        const buttons = [];
        if (isPendingGroup(group)) {
            buttons.push(React.createElement(DynamicButton, { key: `approve${group.ID}`, text: "Approve", color: Color.BLUE, type: ButtonType.BADGE, onClick: approveGroup(group) }));
        }
        buttons.push(React.createElement(Button, { key: `edit${group.ID}`, text: "Edit", color: Color.YELLOW, type: ButtonType.BADGE, className: "ml-2", onClick: handleEditGroup(group) }));
        buttons.push(React.createElement(DynamicButton, { key: `delete${group.ID}`, text: "Delete", color: Color.RED, type: ButtonType.BADGE, className: "ml-2", onClick: handleDeleteGroup(group) }));
        return React.createElement("td", { className: "d-flex" }, buttons);
    };
    const GroupMembers = ({ group }) => {
        if (!hasUsers(group)) {
            return React.createElement("td", null, "No members");
        }
        const members = group.users.map((user, index) => {
            return (React.createElement("span", { key: user.ID.toString(), className: "inline-block" },
                React.createElement("a", { href: `https://github.com/${user.Login}`, target: "_blank", rel: "noopener noreferrer" }, user.Name),
                index >= group.users.length - 1 ? "" : ", "));
        });
        return React.createElement("td", null, members);
    };
    const GroupRow = ({ group }) => {
        return (React.createElement("tr", { hidden: groupSearch(group) },
            React.createElement("td", { key: group.ID.toString() },
                group.name,
                React.createElement("span", { className: "badge badge-warning ml-2" }, isPendingGroup(group) ? "Pending" : null)),
            React.createElement(GroupMembers, { group: group }),
            React.createElement(GroupButtons, { group: group })));
    };
    const PendingGroups = state.groups[courseID.toString()]?.filter(group => isPendingGroup(group)).map(group => {
        return React.createElement(GroupRow, { key: group.ID.toString(), group: group });
    });
    const ApprovedGroups = state.groups[courseID.toString()]?.filter(group => isApprovedGroup(group)).map(group => {
        return React.createElement(GroupRow, { key: group.ID.toString(), group: group });
    });
    if (state.activeGroup) {
        return React.createElement(GroupForm, { key: state.activeGroup.ID.toString() });
    }
    const table = (React.createElement("table", { className: "table table-striped table-grp table-hover" },
        React.createElement("thead", { className: "thead-dark" },
            React.createElement("tr", null,
                React.createElement("th", null, "Name"),
                React.createElement("th", null, "Members"),
                React.createElement("th", null, "Manage"))),
        React.createElement("tbody", null,
            PendingGroups,
            ApprovedGroups)));
    return (React.createElement("div", { className: "box" },
        React.createElement("div", { className: "pb-2" },
            React.createElement(Search, null)),
        table));
};
export default Groups;
