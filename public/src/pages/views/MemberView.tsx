import * as React from "react";
import { Course, Enrollment, Status, Users } from '../../../proto/ag_pb';
import { Search } from "../../components";
import { CourseManager, ILink, NavigationManager } from "../../managers";
import { IUserRelation } from "../../models";
import { ActionType, UserView } from "./UserView";

interface IUserViewerProps {
    navMan: NavigationManager;
    courseMan: CourseManager;
    acceptedUsers: IUserRelation[];
    pendingUsers: IUserRelation[];
    course: Course;
    courseURL: string;
}

interface IUserViewerState {
    acceptedUsers: IUserRelation[];
    pendingUsers: IUserRelation[];
    approveAllClicked: boolean;
    editing: boolean;
    pendingUsersView?: JSX.Element;
    acceptedUsersView?: JSX.Element;
    errMsg: JSX.Element | null;
}

export class MemberView extends React.Component<IUserViewerProps, IUserViewerState> {

    constructor(props: IUserViewerProps) {
        super(props);
        this.state = {
            acceptedUsers: this.props.acceptedUsers,
            pendingUsers: this.props.pendingUsers,
            approveAllClicked: false,
            editing: false,
            errMsg: null,
        };
    }
    public render() {
        return <div>
            <h1>{this.props.course.getName()}</h1>
            <Search className="input-group"
                    placeholder="Search for users"
                    onChange={(query) => this.handleSearch(query)}
                />
            {this.state.errMsg}
            {this.state.pendingUsersView}
            {this.state.acceptedUsersView}
        </div>;
    }

    public componentWillMount() {
        this.setState({
            pendingUsersView: this.renderPendingView(),
            acceptedUsersView: this.renderUserView(),
        });
    }

    public renderUserView() {
        const header = <div> Registered users {this.editButton()}</div>;
        if (this.state.acceptedUsers.length > 0 || this.props.acceptedUsers.length > 0) {
            return this.renderUsers(
                header,
                this.state.acceptedUsers,
                [],
                ActionType.InRow,
                (user: IUserRelation) => {
                    return this.generateUserButtons(user.enrollment);
                });
        }
    }

    public renderPendingView() {
        if (this.props.pendingUsers.length > 0 || this.state.pendingUsers.length > 0) {
            const header = <div> Pending users {this.approveButton()}</div>;
            return this.renderUsers(
                header,
                this.state.pendingUsers,
                [],
                ActionType.InRow,
                (user: IUserRelation) => {
                    return this.generateUserButtons(user.enrollment);
                });
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
                await this.handleAccept(userRel);
                break;
            case "reject":
                await this.handleReject(userRel);
                break;
            case "teacher":
                await this.handlePromote(userRel);
                break;
            case "demote":
                await this.handleDemote(userRel);
                break;
        }
        this.toggleEditState();
        this.props.navMan.refresh();
        this.toggleEditState();
    }

    private async handleAccept(userRel: IUserRelation) {
        const result = await this.props.courseMan.changeUserState(userRel.enrollment, Enrollment.UserStatus.STUDENT);
        this.checkForErrors(result, () => {
            userRel.enrollment.setStatus(Enrollment.UserStatus.STUDENT);
            const i = this.state.pendingUsers.indexOf(userRel);
            if (i >= 0) {
                this.state.pendingUsers.splice(i, 1);
                this.state.acceptedUsers.push(userRel);
            }
        })
    }

    private async handleReject(userRel: IUserRelation) {
        if (confirm(
            `Warning! This action is irreversible!
            Do you want to reject the student?`,
        )) {
            let readyToDelete =
             await this.props.courseMan.isEmptyRepo(this.props.course.getId(), userRel.user.getId(), 0);
            if (!readyToDelete) {
                readyToDelete = confirm(
                    `Warning! User repository is not empty.
                    Do you still want to reject the user?`,
                    );
            }

            if (readyToDelete) {
                const result =
            await this.props.courseMan.changeUserState(userRel.enrollment, Enrollment.UserStatus.NONE);
                this.checkForErrors(result, () => {
                    const i = this.state.pendingUsers.indexOf(userRel);
                    if (i >= 0) {
                        this.state.pendingUsers.splice(i, 1);
                    }
                    const j = this.state.acceptedUsers.indexOf(userRel);
                    if (j >= 0) {
                        this.state.acceptedUsers.splice(j, 1);
                    }
                })
            }
       }
    }

    private async handlePromote(userRel: IUserRelation) {
        if (confirm(
            `Are you sure you want to promote
            ${userRel.user.getName()} to teacher status?`,
        )) {
            const result = await this.props.courseMan.changeUserState(userRel.enrollment, Enrollment.UserStatus.TEACHER);
            this.checkForErrors(result);
        }
        this.toggleEditState();
    }

    private async handleDemote(userRel: IUserRelation) {
        if (confirm(
            `Warning! ${userRel.user.getName()} is a teacher.
            Do you want to demote ${userRel.user.getName()} to student?`,
        )) {
            const result = await this.props.courseMan.changeUserState(userRel.enrollment, Enrollment.UserStatus.STUDENT);
            this.checkForErrors(result);
        }
        this.toggleEditState();
    }

    private handleSearch(query: string): void {
        query = query.toLowerCase();
        const filteredAccepted: IUserRelation[] = [];
        const filteredPending: IUserRelation[] = [];

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

    private editButton() {
        return <button type="button"
                id="edit"
                className="btn btn-success member-btn"
                onClick={() => this.toggleEditState()}
    >{this.editButtonString()}</button>;
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

    private toggleEditState() {
        this.setState({
            editing: !this.state.editing,

        }, () => this.setState({
            pendingUsersView: this.renderPendingView(),
            acceptedUsersView: this.renderUserView(),
        }));
    }

    private editButtonString(): string {
        return this.state.editing ? "Cancel" : "Edit";
    }

    private generateUserButtons(enrollment: Enrollment): ILink[] {
        const links = [];
        switch (enrollment.getStatus()) {
            case Enrollment.UserStatus.PENDING:
                links.push({
                    name: "Accept",
                    extra: "primary",
                    uri: "accept",
                }, {
                    name: "Reject",
                    extra: "danger",
                    uri: "reject",
                });
                break;
            case Enrollment.UserStatus.STUDENT:
                this.state.editing ? links.push({
                    name: "Promote",
                    extra: "primary",
                    uri: "teacher",
                }, {
                    name: "Reject",
                    extra: "danger",
                    uri: "reject",
                }) : links.push({
                    name: "Student",
                    extra: "light",
                });
                break;
            case Enrollment.UserStatus.TEACHER:
                this.state.editing ? links.push({
                    name: "Demote ",
                    extra: "primary",
                    uri: "demote",
                }, {
                    name: "Reject",
                    extra: "danger",
                    uri: "reject",
                }) : links.push({
                    name: "Teacher",
                    extra: "light",
                });
                break;
            default:
                console.log("Got unexpected user status " + enrollment.getStatus() + " when generating links");
        }
        return links;
    }

    private checkForErrors(status: Status, action?: () => void) {
        if (status.getCode() !== 0) {
            this.generateErrorMessage(status);
            return;
        } else if (action) {
            action();
        }
    }

    private generateErrorMessage(status: Status) {
        const err = <div className="alert alert-danger">{status.getError()}</div>;
        this.setState({
                errMsg: err,
        });
    }
}
