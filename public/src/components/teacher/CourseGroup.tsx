import * as React from "react";
import { Course, Group, User } from "../../../proto/ag_pb";
import { BootstrapButton, DynamicTable, Search } from "../../components";
import { LiDropDownMenu } from "../../components/navigation/LiDropDownMenu";
import { bindFunc, RProp } from "../../helper";
import { CourseManager, ILink, NavigationManager } from "../../managers";
import { BootstrapClass } from "../bootstrap/BootstrapButton";
import { generateGroupRepoLink } from "./groupHelper";

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
}

export class CourseGroup extends React.Component<ICourseGroupProps, ICourseGroupState> {

    constructor(props: any) {
        super(props);
        this.state = {
            approvedGroups: this.props.approvedGroups,
            pendingGroups: this.props.pendingGroups,
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
                {noGroupsWell}
                {approvedGroups}
                {pendingGroups}
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
                <h3>Approved Groups</h3>
                <DynamicTable
                    header={["Name", "Members"]}
                    data={this.state.approvedGroups}
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
                    selector={(group: Group) => this.renderRow(group, false)}
                />
            </div>
        );
    }

    private renderRow(group: Group, withLink: boolean): Array<string | JSX.Element> {
        const selector: Array<string | JSX.Element> = [];
        const groupName = withLink ? generateGroupRepoLink(group.getName(), this.props.courseURL) : group.getName();
        selector.push(groupName, this.getMembers(group.getUsersList()));
        const dropdownMenu = this.renderDropdownMenu(group);
        selector.push(dropdownMenu);
        return selector;
    }

    private renderDropdownMenu(group: Group): JSX.Element {
        const links = [];
        // only add approve link to not approved groups
        if (group.getStatus() !== Group.GroupStatus.APPROVED) {
            links.push({ name: "Approve", uri: "approve", extra: "primary" });
        }
        links.push({ name: "Edit", uri: "edit", extra: "primary" });
        links.push({ name: "Delete", uri: "delete", extra: "danger" });
        return <ul className="nav nav-pills">
            <LiDropDownMenu
                links={links}
                onClick={(link) => this.handleActionOnClick(group, link)}>
                <span className="glyphicon glyphicon-option-vertical" />
            </LiDropDownMenu>
        </ul>;
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

    private getMembers(users: User[]): JSX.Element {
        const names: JSX.Element[] = [];
        users.forEach((user, i) => {
            let separator = ", ";
            if (i >= users.length - 1) {
                separator = " ";
            }

            const nameLink = <span><a href={this.props.courseURL
                 + user.getLogin() + "-labs"} target="_blank">{ user.getName() }</a>{separator}</span>;
            names.push(nameLink);
            });
        return <div>{names}</div>;
    }

    private async handleUpdateStatus(group: Group, status: Group.GroupStatus): Promise<void> {
        await this.props.courseMan.updateGroupStatus(group.getId(), status);
        this.props.navMan.refresh();
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
            if (ans) {
                this.props.navMan.refresh();
            }
        }
    }

    private async handleActionOnClick(group: Group, link: ILink): Promise<void> {
        switch (link.uri) {
            case "approve":
                group.setStatus(Group.GroupStatus.APPROVED);
                await this.props.courseMan.updateGroup(group);
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

    private refreshState() {
        this.setState({
            approvedGroups: this.props.approvedGroups,
            pendingGroups: this.props.pendingGroups,
        });
        return this.forceUpdate();
    }

}
