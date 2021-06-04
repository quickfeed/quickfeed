import * as React from "react";
import { Assignment, Course, Enrollment, Status } from "../../../proto/ag/ag_pb";
import { Search } from "../../components";
import { searchForStudents } from '../../componentHelper';
import { CourseManager, ILink, NavigationManager } from "../../managers";
import { ActionType, UserView } from "./UserView";

interface IUserViewerProps {
    navMan: NavigationManager;
    courseMan: CourseManager;
    acceptedUsers: Enrollment[];
    pendingUsers: Enrollment[];
    course: Course;
    assignments: Assignment[];
    courseURL: string;
}

interface IUserViewerState {
    acceptedUsers: Enrollment[];
    pendingUsers: Enrollment[];
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
                true,
                [],
                ActionType.InRow,
                (user: Enrollment) => {
                    return this.generateUserButtons(user);
                });
        }
    }

    public renderPendingView() {
        if (this.props.pendingUsers.length > 0 || this.state.pendingUsers.length > 0) {
            const header = <div> Pending users {this.approveButton()}</div>;
            return this.renderUsers(
                header,
                this.state.pendingUsers,
                false,
                [],
                ActionType.InRow,
                (enrollment: Enrollment) => {
                    return this.generateUserButtons(enrollment);
                });
        }
    }

    public renderUsers(
        title: string | JSX.Element,
        enrollments: Enrollment[],
        withActivity: boolean,
        actions?: ILink[],
        linkType?: ActionType,
        optionalActions?: ((enrollment: Enrollment) => ILink[])) {
        return <div>
            <h3>{title}</h3>
            <UserView
                users={enrollments}
                assignments={this.props.assignments}
                actions={actions}
                withActivity={withActivity}
                isCourseList={true}
                courseURL={this.props.courseURL}
                optionalActions={optionalActions}
                linkType={linkType}
                actionClick={(user, link) => this.handleAction(user, link)}
            >
            </UserView>
        </div>;
    }

    public async handleAction(enrol: Enrollment, link: ILink) {
        switch (link.uri) {
            case "accept":
                await this.handleAccept(enrol);
                break;
            case "reject":
                await this.handleReject(enrol);
                break;
            case "teacher":
                await this.handlePromote(enrol);
                break;
            case "demote":
                await this.handleDemote(enrol);
                break;
        }
        this.rerender();
    }

    private async handleAccept(enrol: Enrollment) {
        const result = await this.props.courseMan.changeUserStatus(enrol, Enrollment.UserStatus.STUDENT);
        this.checkForErrors(result, () => {
            enrol.setStatus(Enrollment.UserStatus.STUDENT);
            const i = this.state.pendingUsers.indexOf(enrol);
            if (i >= 0) {
                this.state.pendingUsers.splice(i, 1);
                this.state.acceptedUsers.push(enrol);
            }
        })
    }

    private async handleReject(enrol: Enrollment) {
        if (confirm(
            `Warning! This action is irreversible!
            Do you want to reject the student?`,
        )) {
            let readyToDelete =
             await this.props.courseMan.isEmptyRepo(this.props.course.getId(), enrol.getUserid(), 0);
            if (!readyToDelete) {
                readyToDelete = confirm(
                    `Warning! User repository is not empty.
                    Do you still want to reject the user?`,
                    );
            }

            if (readyToDelete) {
                const result =
            await this.props.courseMan.changeUserStatus(enrol, Enrollment.UserStatus.NONE);
                this.checkForErrors(result, () => {
                    const i = this.state.pendingUsers.indexOf(enrol);
                    if (i >= 0) {
                        this.state.pendingUsers.splice(i, 1);
                    }
                    const j = this.state.acceptedUsers.indexOf(enrol);
                    if (j >= 0) {
                        this.state.acceptedUsers.splice(j, 1);
                    }
                })
            }
       }
    }

    private async handlePromote(enrol: Enrollment) {
        if (confirm(
            `Are you sure you want to promote
            ${enrol.getUser()?.getName()} to teacher status?`,
        )) {
            const result = await this.props.courseMan.changeUserStatus(enrol, Enrollment.UserStatus.TEACHER);
            this.checkForErrors(result);
        }
    }

    private async handleDemote(enrol: Enrollment) {
        if (confirm(
            `Warning! ${enrol.getUser()?.getName()} is a teacher.
            Do you want to demote ${enrol.getUser()?.getName()} to student?`,
        )) {
            const result = await this.props.courseMan.changeUserStatus(enrol, Enrollment.UserStatus.STUDENT);
            this.checkForErrors(result);
        }
    }

    private handleSearch(query: string) {
        this.setState({
            acceptedUsers: searchForStudents(this.props.acceptedUsers, query),
            pendingUsers: searchForStudents(this.props.pendingUsers, query),
        }, () => this.rerender());
    }

    private approveButton() {
        return <button type="button"
                id="approve"
                className="btn btn-success member-btn"
                // only activate the approve function if is not already approving
                onClick={this.state.approveAllClicked ?
                    () => {return; } : async () => {
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

        }, () => this.rerender());
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
        this.clearErrorMessage();
    }

    private generateErrorMessage(status: Status) {
        const err = <div className="alert alert-danger">{status.getError()}</div>;
        this.setState({
                errMsg: err,
        });
    }

    private clearErrorMessage() {
        this.setState({
            errMsg: <div></div>,
        });
    }

    private rerender() {
        this.setState({
            pendingUsersView: this.renderPendingView(),
            acceptedUsersView: this.renderUserView(),
        });
    }
}
