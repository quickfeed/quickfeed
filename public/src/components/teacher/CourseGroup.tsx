import * as React from "react";

import { DynamicTable } from "../../components";
import { CourseManager, NavigationManager } from "../../managers";
import { CourseGroupStatus, ICourse, ICourseGroup, IUser } from "../../models";

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
                                <button
                                    onClick={(e) => this.handleUpdateStatus(group.id, CourseGroupStatus.approved)}
                                    className="btn btn-primary">
                                    Approve
                        </button>
                                <button
                                    onClick={(e) => this.handleUpdateStatus(group.id, CourseGroupStatus.rejected)}
                                    className="btn btn-danger"> Reject
                    </button>
                            </span>,
                        ]}
                />
            </div>
        );
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
