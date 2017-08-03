import * as React from "react";

import { BootstrapButton, DynamicTable } from "../../components";
import { CourseManager, NavigationManager } from "../../managers";
import { CourseGroupStatus, ICourse, ICourseGroup, IUser } from "../../models";

import { bindFunc, RProp } from "../../helper";
import { BootstrapClass } from "../bootstrap/BootstrapButton";

interface ICourseGroupProp {
    approvedGroups: ICourseGroup[];
    pendingGroups: ICourseGroup[];
    course: ICourse;
    navMan: NavigationManager;
    courseMan: CourseManager;
}

class CourseGroup extends React.Component<ICourseGroupProp, any> {
    public render() {
        let approvedGroups;
        if (this.props.approvedGroups.length > 0) {
            approvedGroups = this.createApproveGroupView();
        }
        let pendingGroups;
        if (this.props.pendingGroups.length > 0) {
            pendingGroups = this.createPendingGroupView();
        }
        return (
            <div className="group-container">
                <h1>{this.props.course.name}</h1>
                {approvedGroups}
                {pendingGroups}
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
                    header={["Name", "Members", "Action"]}
                    data={this.props.pendingGroups}
                    selector={
                        (group: ICourseGroup) => [
                            group.name,
                            this.getMembers(group.users),
                            <span>
                                <UpdateButton type="primary" group={group} status={CourseGroupStatus.approved}>
                                    Approve
                                </UpdateButton>
                                <UpdateButton type="danger" group={group} status={CourseGroupStatus.rejected}>
                                    Reject
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
        console.log(props);
        return <BootstrapButton
            onClick={(e) => this.handleUpdateStatus(props.group.id, props.status)}
            classType={props.type}>>
            {props.children}
        </BootstrapButton>;
    }

    private getMembers(users: IUser[]): string {
        const names: string[] = [];
        for (const user of users) {
            // names.push(user.firstname + " " + user.lastname);
            names.push(user.id.toString());
        }
        return names.toString();
    }

    private async handleUpdateStatus(gid: number, status: CourseGroupStatus): Promise<void> {
        const result = await this.props.courseMan.updateGroupStatus(gid, status);
        if (result) {
            this.props.navMan.refresh();
        }
    }
}

export { CourseGroup };
