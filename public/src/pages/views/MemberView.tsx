import * as React from "react";
import { Course, Enrollment } from "../../../proto/ag_pb";
import { Search } from "../../components";
import { CourseManager, ILink, NavigationManager } from "../../managers";
import { IUserRelation } from "../../models";
import { ActionType, UserView } from "./UserView";

interface IUserViewerProps {
    navMan: NavigationManager;
    courseMan: CourseManager;
    acceptedUsers: IUserRelation[];
    pendingUsers: IUserRelation[];
    rejectedUsers: IUserRelation[];
    course: Course;
    courseURL: string;
}

interface IUserViewerState {
    acceptedUsers: IUserRelation[];
    pendingUsers: IUserRelation[];
    rejectedUsers: IUserRelation[];
    approveAllClicked: boolean;
}

export class MemberView extends React.Component<IUserViewerProps, IUserViewerState> {

    constructor(props: IUserViewerProps) {
        super(props);
        this.state = {
            acceptedUsers: this.props.acceptedUsers,
            pendingUsers: this.props.pendingUsers,
            rejectedUsers: this.props.rejectedUsers,
            approveAllClicked: false,
        };
    }
    public render() {
        const pendingActions = [
            { name: "Accept", uri: "accept", extra: "primary" },
            { name: "Reject", uri: "reject", extra: "danger" },
        ];
        return <div>
            <h1>{this.props.course.getName()}</h1>
            <Search className="input-group"
                    placeholder="Search for users"
                    onChange={(query) => this.handleSearch(query)}
                />
            {this.renderPendingView(pendingActions)}
            {this.renderUserView()}
            {this.renderRejectedView()}
        </div>;
    }

    public componentWillReceiveProps(newProps: IUserViewerProps) {
        this.setState({
            acceptedUsers: newProps.acceptedUsers,
            rejectedUsers: newProps.rejectedUsers,
            pendingUsers: newProps.pendingUsers,
        });
    }

    public renderRejectedView() {
        if (this.state.rejectedUsers.length > 0 || this.props.rejectedUsers.length > 0) {
            return this.renderUsers(
                "Rejected users",
                this.state.rejectedUsers,
                [],
                ActionType.Menu,
                (user: IUserRelation) => {
                    const links = [];
                    if (user.link.getStatus() === Enrollment.UserStatus.REJECTED) {
                        links.push({ name: "Set pending", uri: "remove", extra: "primary" });
                    }
                    return links;
                });
        }
    }

    public renderUserView() {
        if (this.state.acceptedUsers.length > 0 || this.props.acceptedUsers.length > 0) {
            return this.renderUsers(
                "Registered users",
                this.state.acceptedUsers,
                [],
                ActionType.InRow,
                (user: IUserRelation) => {
                    const links = [];
                    if (user.link.getStatus() === Enrollment.UserStatus.TEACHER) {
                        links.push({ name: "Teacher", extra: "light" });
                    } else {
                        links.push({ name: "Assign as Teacher", uri: "teacher", extra: "primary" });
                        links.push({ name: "Remove", uri: "reject", extra: "danger" });
                    }
                    return links;
                });
        }
    }

    public renderPendingView(pendingActions: ILink[]) {
        if (this.props.pendingUsers.length > 0 || this.state.pendingUsers.length > 0) {
            const header = <div> Pending users {this.approveButton()}</div>;
            return this.renderUsers(header, this.state.pendingUsers, pendingActions, ActionType.InRow);
        }
    }

    public renderUsers(
        title: string | JSX.Element,
        users: IUserRelation[],
        actions?: ILink[],
        linkType?: ActionType,
        optionalActions?: ((user: IUserRelation) => ILink[])) {
        return <div>
            <h3>{title}</h3>
            <UserView
                users={users}
                actions={actions}
                isCourseList={true}
                courseURL={this.props.courseURL}
                optionalActions={optionalActions}
                linkType={linkType}
                actionClick={(user, link) => this.handleAction(user, link)}
            >
            </UserView>
        </div>;
    }

    public async handleAction(userRel: IUserRelation, link: ILink) {
        switch (link.uri) {
            case "accept":
                this.handleAccept(userRel);
                break;
            case "reject":
                this.handleReject(userRel);
                break;
            case "teacher":
                this.handlePromote(userRel);
                break;
            case "remove":
                this.handleRemove(userRel);
                break;
        }
        this.props.navMan.refresh();
    }

    private async handleAccept(userRel: IUserRelation) {
        const result = await this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.STUDENT);
        if (result) {
            userRel.link.setStatus(Enrollment.UserStatus.STUDENT);
            const i = this.state.pendingUsers.indexOf(userRel);
            if (i >= 0) {
                this.state.pendingUsers.splice(i, 1);
                this.state.acceptedUsers.push(userRel);
            }
        }
    }

    private async handleReject(userRel: IUserRelation) {
        if (confirm(
            `Warning! This action is irreversible!
            Do you want to reject the student?`,
        )) {
            const result =
             await this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.REJECTED);
            if (result) {
                userRel.link.setStatus(Enrollment.UserStatus.REJECTED);
            }
        }
    }

    private async handlePromote(userRel: IUserRelation) {
        if (confirm(
            `Warning! This action is irreversible!
            Do you want to continue assigning:
            ${userRel.user.getName()} as a teacher?`,
        )) {
            this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.TEACHER);
        }
    }

    private async handleRemove(userRel: IUserRelation) {
        const result = await this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.PENDING);
        if (result) {
            userRel.link.setStatus(Enrollment.UserStatus.PENDING);
        }
    }

    private handleSearch(query: string): void {
        query = query.toLowerCase();
        const filteredAccepted: IUserRelation[] = [];
        const filteredPending: IUserRelation[] = [];
        const filteredRejected: IUserRelation[] = [];

        // we filter out every student group separately to ensure that student status is easily visible to teacher
        // filter accepted students
        this.props.acceptedUsers.forEach((user) => {
            if (this.found(query, user)) {
                filteredAccepted.push(user);
            }
        });

        this.setState({
            acceptedUsers: filteredAccepted,
        });

        // filter pending students
        this.props.pendingUsers.forEach((user) => {
            if (this.found(query, user)) {
                filteredPending.push(user);
            }
        });

        this.setState({
            pendingUsers: filteredPending,
        });

        // filter rejected students
        this.props.rejectedUsers.forEach((user) => {
            if (this.found(query, user)) {
                filteredRejected.push(user);
            }
        });

        this.setState({
            rejectedUsers: filteredRejected,
        });

    }

    private found(query: string, user: IUserRelation): boolean {
        if (user.user.getName().toLowerCase().indexOf(query) !== -1
                || user.user.getStudentid().toLowerCase().indexOf(query) !== -1
                || user.user.getLogin().toLowerCase().indexOf(query) !== -1
            ) {
                return true;
            }
        return false;
    }

    private approveButton() {
        return <button type="button"
                id="approve"
                className="btn btn-success member-btn"
                // only activate the approve function if is not already approving
                onClick={this.state.approveAllClicked ?
                    () => {} : async () => {
                        await this.handleApproveClick().then(() => {
                            this.setState({approveAllClicked: false});
                        });
                    }
                }> {this.approveButtonString()} </button>;
    }

    private async handleApproveClick(): Promise<boolean> {
        this.setState({approveAllClicked: true});
        const ans = await this.props.courseMan.approveAll(this.props.course.getId());
        this.props.navMan.refresh();
        return ans;
    }

    private approveButtonString(): string {
        return this.state.approveAllClicked ? "Approving..." : "Approve all ";
    }
}
