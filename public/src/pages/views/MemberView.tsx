import * as React from "react";
import { Course, Enrollment, Repository } from "../../../proto/ag_pb";
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
            {this.props.pendingUsers.length > 0 ? this.approveButton() : null}
            {this.renderUserView()}
            {this.renderRejectedView()}
        </div>;
    }

    public componentDidUpdate(prevProps: IUserViewerProps) {
        if ((prevProps.acceptedUsers.length !== this.props.acceptedUsers.length)
         || (prevProps.pendingUsers.length !== this.props.pendingUsers.length)
          || (prevProps.rejectedUsers.length !== this.props.rejectedUsers.length)) {
            return this.refreshState();
       }
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
                ActionType.Menu,
                (user: IUserRelation) => {
                    const links = [];
                    if (user.link.getStatus() === Enrollment.UserStatus.TEACHER) {
                        links.push({ name: "This is a teacher", extra: "primary" });
                    } else {
                        links.push({ name: "Make Teacher", uri: "teacher", extra: "primary" });
                        links.push({ name: "Reject", uri: "reject", extra: "danger" });
                    }
                    return links;
                });
        }
    }

    public renderPendingView(pendingActions: ILink[]) {
        if (this.props.pendingUsers.length > 0 || this.state.pendingUsers.length > 0) {
            return this.renderUsers("Pending users", this.state.pendingUsers, pendingActions, ActionType.InRow);
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
                isCourseList={true}
                courseURL={this.props.courseURL}
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
                this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.STUDENT).then((ans) => {
                    console.log("List of accepted users updated: " + ans);
                });
                break;
            case "reject":
                this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.REJECTED);
                break;
            case "teacher":
                if (confirm(
                    `Warning! This action is irreversible!
                    Do you want to continue assigning:
                    ${userRel.user.getName()} as a teacher?`,
                )) {
                    this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.TEACHER);
                }
                break;
            case "remove":
                this.props.courseMan.changeUserState(userRel.link, Enrollment.UserStatus.PENDING);
                break;
        }
        this.refreshState();
        this.props.navMan.refresh();
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
        return <p> <button type="button"
                id="approve"
                className="btn btn-success"
                // only activate the approve function if is not already approving
                onClick={this.state.approveAllClicked ?
                    () => {} : async () => {
                        await this.handleApproveClick().then(() => {
                            this.setState({approveAllClicked: false});
                        });
                    }
                }> {this.approveButtonString()} </button> </p>;
    }

    private async handleApproveClick(): Promise<boolean> {
        this.setState({approveAllClicked: true});
        const ans = await this.props.courseMan.approveAll(this.props.course.getId());
        console.log("ApproveAll got answer: " + ans);
        this.props.navMan.refresh();
        return ans;
    }

    private approveButtonString(): string {
        return this.state.approveAllClicked ? "Approving..." : "Approve all pending";
    }

    private refreshState() {
        this.setState({
            pendingUsers: this.props.pendingUsers,
            acceptedUsers: this.props.acceptedUsers,
            rejectedUsers: this.props.rejectedUsers,
        });
        return this.forceUpdate();
    }
}
