import * as React from "react";
import { Course, Group, User, Status } from '../../../proto/ag_pb';
import { BootstrapButton, DynamicTable, Search } from "../../components";
import { bindFunc, RProp, generateLabRepoLink } from '../../helper';
import { CourseManager, ILink, NavigationManager } from "../../managers";
import { BootstrapClass } from "../bootstrap/BootstrapButton";
import { generateGroupRepoLink } from "./labHelper";

interface ICourseGroupProps {
    approvedGroups: Group[];
    pendingGroups: Group[];
    course: Course;
    courseURL: string;
    navMan: NavigationManager;
    courseMan: CourseManager;
    pagePath: string;
}

interface ICourseGroupState {
    approvedGroups: Group[];
    pendingGroups: Group[];
    editing: boolean;
    errorMsg: JSX.Element | null;
}

export class CourseGroup extends React.Component<ICourseGroupProps, ICourseGroupState> {

    constructor(props: any) {
        super(props);
        this.state = {
            approvedGroups: this.props.approvedGroups,
            pendingGroups: this.props.pendingGroups,
            editing: false,
            errorMsg: null,
        };
    }

    public render() {
        let approvedGroups;
        if (this.props.approvedGroups.length > 0 || this.state.approvedGroups.length > 0) {
            approvedGroups = this.createApproveGroupView();
        }
        let pendingGroups;
        if (this.props.pendingGroups.length > 0 || this.state.pendingGroups.length > 0) {
            pendingGroups = this.createPendingGroupView();
        }
        let noGroupsWell;
        if (!approvedGroups && !pendingGroups) {
            noGroupsWell = <p className="well">No groups to show!</p>;
        }
        return (
            <div className="group-container">
                <h1>{this.props.course.getName()}</h1>
                <Search className="input-group"
                    placeholder="Search for groups"
                    onChange={(query) => this.handleSearch(query)}
                />
                {this.state.errorMsg}
                {noGroupsWell}
                {pendingGroups}
                {approvedGroups}
            </div>
        );
    }

    public componentDidUpdate(prevProps: ICourseGroupProps) {
        if ((prevProps.approvedGroups.length !== this.props.approvedGroups.length)
            || (prevProps.pendingGroups.length !== this.props.pendingGroups.length)) {
            return this.refreshState();
        }
    }

    private createApproveGroupView(): JSX.Element {
        return (
            <div className="approved-groups">
                <h3>Approved Groups</h3> {this.editButton()}
                <DynamicTable
                    header={["Name", "Members", "Status"]}
                    data={this.state.approvedGroups}
                    classType={"table-grp"}
                    selector={
                        (group: Group) => this.renderRow(group, true)
                    }
                />
            </div>
        );
    }

    private createPendingGroupView(): JSX.Element {
        const UpdateButton = bindFunc(this, this.updateButton);
        return (
            <div className="pending-groups">
                <h3>Pending Groups</h3>
                <DynamicTable
                    header={["Name", "Members", "Actions"]}
                    data={this.state.pendingGroups}
                    classType={"table-grp"}
                    selector={(group: Group) => this.renderRow(group, false)}
                />
            </div>
        );
    }

    private renderRow(group: Group, withLink: boolean): (string | JSX.Element)[] {
        const selector: (string | JSX.Element)[] = [];
        const groupName = withLink ? generateGroupRepoLink(group.getName(), this.props.courseURL) : group.getName();
        selector.push(groupName, this.getMembers(group.getUsersList()));
        const actionButtonLinks = this.generateGroupButtons(group);
        const actionButtons = this.renderActionRow(group, actionButtonLinks);
        selector.push(<div className="btn-group action-btn">{actionButtons}</div>);
        return selector;
    }

    private renderActionRow(group: Group, tempActions: ILink[]) {
        return tempActions.map((v, i) => {
            return <BootstrapButton
                key={i}
                classType={v.extra ? v.extra as BootstrapClass : "default"}
                type={v.description}
                onClick={(link) => { this.handleActionOnClick(group, v)}}
            >{v.name}
            </BootstrapButton>;
        });
    }

    private updateButton(props: RProp<{
        type: BootstrapClass,
        group: Group,
        status: Group.GroupStatus,
    }>) {
        return <BootstrapButton
            onClick={(e) => this.handleUpdateStatus(props.group, props.status)}
            classType={props.type}>
            {props.children}
        </BootstrapButton>;
    }

    private editButton() {
        return <button type="button"
                id="edit"
                className="btn btn-success member-btn"
                onClick={() => this.toggleEditState()}
        >{this.editButtonString()}</button>;
    }

    private getMembers(users: User[]): JSX.Element {
        const names: JSX.Element[] = [];
        users.forEach((user, i) => {
            let separator = ", ";
            if (i >= users.length - 1) {
                separator = " ";
            }

            const nameLink = <span key={"s" + i} ><a href={ generateLabRepoLink(this.props.courseURL, user.getLogin())}
             target="_blank">{ user.getName() }</a>{separator}</span>;
            names.push(nameLink);
            });
        return <div>{names}</div>;
    }

    private async handleUpdateStatus(group: Group, status: Group.GroupStatus): Promise<void> {
        const ans = await this.props.courseMan.updateGroupStatus(group.getId(), status);
        this.checkForErrors(ans);
    }

    private async deleteGroup(group: Group) {
        let readyToDelete = group.getStatus() === Group.GroupStatus.PENDING;
        const courseID = this.props.course.getId();
        // if approved group - check if repo is empty
        if (!readyToDelete) {
            readyToDelete = await this.props.courseMan.isEmptyRepo(courseID, 0, group.getId());

            if (!readyToDelete) {
                readyToDelete = confirm(
                    `Warning! Group repository is not empty!
                    Do you still want to delete group, github team
                    and group repository?`,
                );
            }
        }

        if (readyToDelete) {
            const ans = await this.props.courseMan.deleteGroup(courseID, group.getId());
            this.checkForErrors(ans);
        }
    }

    private async handleActionOnClick(group: Group, link: ILink): Promise<void> {
        switch (link.uri) {
            case "approve":
                const ans = await this.props.courseMan.updateGroup(group);
                this.checkForErrors(ans, () => {
                    group.setStatus(Group.GroupStatus.APPROVED);
                })
                break;
            case "edit":
                this.props.navMan
                    .navigateTo(this.props.pagePath + "/courses/"
                        + group.getCourseid() + "/groups/" + group.getId() + "/edit");
                break;
            case "delete":
                if (confirm(
                    `Warning! This action is irreversible!
                    Do you want to delete group: ${group.getName()}?`,
                )) {
                    await this.deleteGroup(group);
                    break;
                }
        }
        this.refreshState();
        this.props.navMan.refresh();
    }

    private handleSearch(query: string): void {
        query = query.toLowerCase();
        const filteredApproved: Group[] = [];
        const filteredPending: Group[] = [];

        this.props.approvedGroups.forEach((grp) => {
            if (grp.getName().toLowerCase().indexOf(query) !== -1
            ) {
                filteredApproved.push(grp);
            }
        });

        this.setState({
            approvedGroups: filteredApproved,
        });

        this.props.pendingGroups.forEach((grp) => {
            if (grp.getName().toLowerCase().indexOf(query) !== -1
            ) {
                filteredPending.push(grp);
            }
        });

        this.setState({
            pendingGroups: filteredPending,
        });
    }

    private generateGroupButtons(group: Group): ILink[] {
        const links = [];
        switch (group.getStatus()) {
            case Group.GroupStatus.PENDING:
                links.push({
                    name: "Approve",
                    extra: "primary",
                    uri: "approve",
                }, {
                    name: "Edit",
                    extra: "primary",
                    uri: "edit",
                }, {
                    name: "Delete",
                    extra: "danger",
                    uri: "delete",
                });
                break;
            case Group.GroupStatus.APPROVED:
                this.state.editing ? links.push({
                    name: "Edit",
                    extra: "primary",
                    uri: "edit",
                }, {
                    name: "Delete",
                    extra: "danger",
                    uri: "delete",
                }) : links.push({
                    name: "Approved",
                    extra: "light",
                });
                break;
            default:
                console.log("Got unexpected group status " + group.getStatus() + " when generating links");
        }
        return links;
    }

    private toggleEditState() {
        this.setState({
            editing: !this.state.editing,
        }, () => this.refreshState());
    }

    private editButtonString(): string {
        return this.state.editing ? "Cancel" : "Edit";
    }

    private generateErrorMessage(status: Status) {
        const err = <div className="alert alert-danger">{status.getError()}</div>;
        this.setState({
                errorMsg: err,
        });
    }

    private checkForErrors(status: Status, action?: () => void) {
        if (status.getCode() !== 0) {
            this.generateErrorMessage(status);
            return;
        } else if (action) {
            action();
        }
    }

    private refreshState() {
        this.setState({
            approvedGroups: this.props.approvedGroups,
            pendingGroups: this.props.pendingGroups,
            errorMsg: null,
        });
        return this.forceUpdate();
    }
}
