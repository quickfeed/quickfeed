import React, { useCallback, useEffect } from "react";
import { Color, isHidden, userLink } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import DynamicButton from "../DynamicButton";
import DynamicTable from "../DynamicTable";
import Search from "../Search";
import { ButtonType } from "./Button";
import UserComponent from "./User";
const Users = () => {
    const state = useAppState();
    const actions = useActions().global;
    useEffect(() => {
        actions.getUsers();
    }, [actions]);
    const handlePromoteAdmin = useCallback((user) => () => actions.updateAdmin(user), [actions]);
    const headers = ["Name", "GitHub", "Email", "Student ID", "Role"];
    const users = state.allUsers.map((user) => {
        const roleActionText = user.IsAdmin ? "Demote" : "Promote";
        const buttonColor = user.IsAdmin ? Color.RED : Color.BLUE;
        return [
            React.createElement(UserComponent, { key: user.ID, user: user, hidden: !isHidden(user.Name, state.query) }),
            React.createElement("a", { key: `${user.ID}link`, href: userLink(user) }, user.Login),
            user.Email,
            user.StudentID,
            React.createElement(DynamicButton, { key: `${user.ID}button`, text: roleActionText, color: buttonColor, type: ButtonType.BADGE, onClick: handlePromoteAdmin(user) })
        ];
    });
    return (React.createElement("div", null,
        React.createElement("div", { className: "pb-2" },
            React.createElement(Search, null)),
        React.createElement(DynamicTable, { header: headers, data: users })));
};
export default Users;
