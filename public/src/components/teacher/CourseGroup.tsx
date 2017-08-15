import * as React from "react";

import { BootstrapButton, DynamicTable } from "../../components";
import { CourseManager, ILink, NavigationManager } from "../../managers";
import { CourseGroupStatus, ICourse, ICourseGroup, IUser } from "../../models";

import { bindFunc, RProp } from "../../helper";
import { BootstrapClass } from "../bootstrap/BootstrapButton";

import { LiDropDownMenu } from "../../components/navigation/LiDropDownMenu";

interface ICourseGroupProp {
    approvedGroups: ICourseGroup[];
    pendingGroups: ICourseGroup[];
    rejectedGroups: ICourseGroup[];
    course: ICourse;
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
                <h1>{this.props.course.name}</h1>
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
                        (group: ICourseGroup) => [
                            group.name,
                            this.getMembers(group.users),
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
                    selector={(group: ICourseGroup) => this.renderRow(group)}
                />
            </div>
        );
    }

    private renderRow(group: ICourseGroup): Array<string | JSX.Element> {
        const selector: Array<string | JSX.Element> = [];
        selector.push(group.name, this.getMembers(group.users));
        const temp = this.renderDropdownMenu(group);
        selector.push(temp);
        return selector;
    }

    private renderDropdownMenu(group: ICourseGroup): JSX.Element {
        const links = [];
        links.push({ name: "Approve", uri: "approve", extra: "primary" });
        links.push({ name: "Edit", uri: "edit", extra: "primary" });
        links.push({ name: "Reject", uri: "reject", extra: "danger" });
        links.push({ name: "Delete", uri: "delete", extra: "danger" });
        return <ul className="nav nav-pills">
            <LiDropDownMenu
                links={links}
                onClick={(link) => this.handleActionOnClick(group, link)}
            >
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
                        (group: ICourseGroup) => [
                            group.name,
                            this.getMembers(group.users),
                            <span>
                                <UpdateButton type="danger" group={group} status={CourseGroupStatus.deleted}>
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
        group: ICourseGroup,
        status: CourseGroupStatus,
    }>) {
        return <BootstrapButton
            onClick={(e) => this.handleUpdateStatus(props.group.id, props.status)}
            classType={props.type}>
            {props.children}
        </BootstrapButton>;
    }

    private getMembers(users: IUser[]): string {
        const names: string[] = [];
        for (const user of users) {
            names.push(user.name);
        }
        return names.toString();
    }

    private async handleUpdateStatus(gid: number, status: CourseGroupStatus): Promise<void> {
        const result = status === CourseGroupStatus.deleted ?
            await this.props.courseMan.deleteGroup(gid) :
            await this.props.courseMan.updateGroupStatus(gid, status);
        if (result) {
            this.props.navMan.refresh();
        }
    }

    private async handleActionOnClick(group: ICourseGroup, link: ILink): Promise<void> {
        switch (link.uri) {
            case "approve":
                await this.props.courseMan.updateGroupStatus(group.id, CourseGroupStatus.approved);
                break;
            case "reject":
                await this.props.courseMan.updateGroupStatus(group.id, CourseGroupStatus.rejected);
                break;
            case "edit":
                this.props.navMan
                    .navigateTo(this.props.pagePath + "/courses/" + group.courseid + "/groups/" + group.id + "/edit");
                break;
            case "delete":
                if (confirm(
                    `Warning! This action is irreversible!

Do you want to delete group:
${group.name}?`,
                )) {
                    await this.props.courseMan.deleteGroup(group.id);
                    break;
                }
        }
        this.props.navMan.refresh();
    }
}
