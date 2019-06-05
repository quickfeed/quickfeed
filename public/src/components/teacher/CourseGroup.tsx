import * as React from "react";

import { Course, Group, User } from "../../../proto/ag_pb";
import { BootstrapButton, DynamicTable } from "../../components";
import { CourseManager, ILink, NavigationManager } from "../../managers";

import { bindFunc, RProp } from "../../helper";
import { BootstrapClass } from "../bootstrap/BootstrapButton";

import { LiDropDownMenu } from "../../components/navigation/LiDropDownMenu";

interface ICourseGroupProp {
    approvedGroups: Group[];
    pendingGroups: Group[];
    rejectedGroups: Group[];
    course: Course;
    navMan: NavigationManager;
    courseMan: CourseManager;
    pagePath: string;
}

export class CourseGroup extends React.Component<ICourseGroupProp, any> {
    public render() {
        let approvedGroups;
        if (this.props.approvedGroups.length > 0) {
            approvedGroups = this.createApproveGroupView();
        }
        let pendingGroups;
        if (this.props.pendingGroups.length > 0) {
            pendingGroups = this.createPendingGroupView();
        }
        let rejectedGroups;
        if (this.props.rejectedGroups.length > 0) {
            rejectedGroups = this.createRejectedGroupView();
        }
        let noGroupsWell;
        if (!approvedGroups && !pendingGroups && !rejectedGroups) {
            noGroupsWell = <p className="well">No groups to show!</p>;
        }
        return (
            <div className="group-container">
                <h1>{this.props.course.getName()}</h1>
                {noGroupsWell}
                {approvedGroups}
                {pendingGroups}
                {rejectedGroups}
            </div>
        );
    }

    private createApproveGroupView(): JSX.Element {
        return (
            <div className="approved-groups">
                <h3>Approved Groups</h3>
                <DynamicTable
                    header={["Name", "Members"]}
                    data={this.props.approvedGroups}
                    selector={
                        (group: Group) => [
                            group.getName(),
                            this.getMembers(group.getUsersList()),
                        ]}
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
                    data={this.props.pendingGroups}
                    selector={(group: Group) => this.renderRow(group)}
                />
            </div>
        );
    }

    private renderRow(group: Group): Array<string | JSX.Element> {
        const selector: Array<string | JSX.Element> = [];
        selector.push(group.getName(), this.getMembers(group.getUsersList()));
        const dropdownMenu = this.renderDropdownMenu(group);
        selector.push(dropdownMenu);
        return selector;
    }

    private renderDropdownMenu(group: Group): JSX.Element {
        const links = [];
        links.push({ name: "Approve", uri: "approve", extra: "primary" });
        links.push({ name: "Edit", uri: "edit", extra: "primary" });
        links.push({ name: "Reject", uri: "reject", extra: "danger" });
        links.push({ name: "Delete", uri: "delete", extra: "danger" });
        return <ul className="nav nav-pills">
            <LiDropDownMenu
                links={links}
                onClick={(link) => this.handleActionOnClick(group, link)}>
                <span className="glyphicon glyphicon-option-vertical" />
            </LiDropDownMenu>
        </ul>;
    }

    private createRejectedGroupView(): JSX.Element {
        const UpdateButton = bindFunc(this, this.updateButton);
        return (
            <div className="pending-groups">
                <h3>Rejected Groups</h3>
                <DynamicTable
                    header={["Name", "Members", "Action"]}
                    data={this.props.rejectedGroups}
                    selector={
                        (group: Group) => [
                            group.getName(),
                            this.getMembers(group.getUsersList()),
                            <span>
                                <UpdateButton type="danger" group={group} status={Group.GroupStatus.DELETED}>
                                    Remove
                                </UpdateButton>
                            </span>,
                        ]}
                />
            </div>
        );
    }

    private updateButton(props: RProp<{
        type: BootstrapClass,
        group: Group,
        status: Group.GroupStatus,
    }>) {
        return <BootstrapButton
            onClick={(e) => this.handleUpdateStatus(props.group.getId(), props.status)}
            classType={props.type}>
            {props.children}
        </BootstrapButton>;
    }

    private getMembers(users: User[]): string {
        const names: string[] = [];
        for (const user of users) {
            names.push(user.getName());
        }
        return names.toString();
    }

    private async handleUpdateStatus(gid: number, status: Group.GroupStatus): Promise<void> {
        const result = status === Group.GroupStatus.DELETED ?
            await this.props.courseMan.deleteGroup(gid) :
            await this.props.courseMan.updateGroupStatus(gid, status);
        if (result) {
            this.props.navMan.refresh();
        }
    }

    private async handleActionOnClick(group: Group, link: ILink): Promise<void> {
        switch (link.uri) {
            case "approve":
                group.setStatus(Group.GroupStatus.APPROVED);
                await this.props.courseMan.updateGroup(group);
                //await this.props.courseMan.updateGroupStatus(group.id, Group.GroupStatus.APPROVED);
                break;
            case "reject":
                group.setStatus(Group.GroupStatus.REJECTED);
                await this.props.courseMan.updateGroup(group);
                break;
            case "edit":
                console.log("CourseGroup: handleActionOnClick attempts to edit group " + group);
                this.props.navMan
                    .navigateTo(this.props.pagePath + "/courses/" + group.getCourseid() + "/groups/" + group.getId() + "/edit");
                break;
            case "delete":
                if (confirm(
                    `Warning! This action is irreversible!

Do you want to delete group:
${group.getName()}?`,
                )) {
                    await this.props.courseMan.deleteGroup(group.getId());
                    break;
                }
        }
        this.props.navMan.refresh();
    }
}
