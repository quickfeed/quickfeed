import * as React from "react";
import { CourseManager, ILink, NavigationManager, UserManager } from "../../managers";
import { CourseUserState, ICourse, ICourseUserLink, IUser, IUserRelation } from "../../models";

import { DynamicTable } from "../../components";
import { ActionType, UserView } from "./UserView";
import { Enrollment } from "../../../proto/ag_pb";

interface IUserViewerProps {
    navMan: NavigationManager;
    courseMan: CourseManager;
    acceptedUsers: IUserRelation[];
    pendingUsers: IUserRelation[];
    rejectedUsers: IUserRelation[];
    course: ICourse;
}

export class MemberView extends React.Component<IUserViewerProps, {}> {
    public render() {
        const pendingActions = [
            { name: "Accept", uri: "accept", extra: "primary" },
            { name: "Reject", uri: "reject", extra: "danger" },
        ];

        return <div>
            <h1>{this.props.course.name}</h1>
            {this.renderUserView()}
            {this.renderPendingView(pendingActions)}
            {this.renderRejectedView()}
        </div>;
    }

    public renderRejectedView() {
        if (this.props.rejectedUsers.length > 0) {
            return this.renderUsers(
                "Rejected users",
                this.props.rejectedUsers,
                [],
                ActionType.Menu,
                (user: IUserRelation) => {
                    const links = [];
                    if (user.link.state === Enrollment.UserStatus.Rejected) {
                        links.push({ name: "Set pending", uri: "remove", extra: "primary" });
                    }
                    return links;
                });
        }
    }

    public renderUserView() {
        return this.renderUsers(
            "Registered users",
            this.props.acceptedUsers,
            [],
            ActionType.Menu,
            (user: IUserRelation) => {
                const links = [];
                if (user.link.state === Enrollment.UserStatus.Teacher) {
                    links.push({ name: "This is a teacher", extra: "primary" });
                } else {
                    links.push({ name: "Make Teacher", uri: "teacher", extra: "primary" });
                    links.push({ name: "Reject", uri: "reject", extra: "danger" });
                }

                return links;
            });
    }

    public renderPendingView(pendingActions: ILink[]) {
        if (this.props.pendingUsers.length > 0) {
            return this.renderUsers("Pending users", this.props.pendingUsers, pendingActions, ActionType.InRow);
        }
    }

    public renderUsers(
        title: string,
        users: IUserRelation[],
        actions?: ILink[],
        linkType?: ActionType,
        optionalActions?: ((user: IUserRelation) => ILink[])) {
        return <div>
            <h3>{title}</h3>
            <UserView
                users={users}
                actions={actions}
                optionalActions={optionalActions}
                linkType={linkType}
                actionClick={(user, link) => this.handleAction(user, link)}
            >
            </UserView>
        </div>;
    }

    public handleAction(userRel: IUserRelation, link: ILink) {
        switch (link.uri) {
            case "accept":
                this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.Student);
                break;
            case "reject":
                this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.Rejected);
                break;
            case "teacher":
                if (confirm(
                    `Warning! This action is irreversible!
                    Do you want to continue assigning:
                    ${userRel.user.getName()} as a teacher?`,
                )) {
                    this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.Teacher);
                }
                break;
            case "remove":
                this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.Pending);
                break;
        }
        this.props.navMan.refresh();
    }
}

export { UserView };
