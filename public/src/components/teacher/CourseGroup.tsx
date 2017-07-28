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

interface ICourseGroupState {
    approvedGroups: ICourseGroup[];
    pendingGroups: ICourseGroup[];
}

class CourseGroup extends React.Component<ICourseGroupProp, ICourseGroupState> {
    constructor(props: any) {
        super(props);
        this.state = {
            approvedGroups: this.props.approvedGroups,
            pendingGroups: this.props.pendingGroups,
        };
    }
    public render() {
        const approvedGroups = this.createApproveGroupView();
        const pendingGroups = this.createPendingGroupView();
        return (
            <div className="group-container">
                <h1>{this.props.course.name}</h1>
                <div className="approved-groups">
                    <h3>Approved Groups</h3>
                    {approvedGroups}
                </div>
                <div className="pending-groups">
                    <h3>Pending Groups</h3>
                    {pendingGroups}
                </div>
            </div>
        );
    }

    private createApproveGroupView(): JSX.Element {
        return (
            <DynamicTable
                header={["Name", "Members"]}
                data={this.state.approvedGroups}
                selector={
                    (group: ICourseGroup) => [
                        group.name,
                        this.getMembers(group.users),
                    ]}
            />
        );
    }

    private createPendingGroupView(): JSX.Element {
        return (
            <DynamicTable
                header={["Name", "Members", "Action"]}
                data={this.state.pendingGroups}
                selector={
                    (group: ICourseGroup) => [
                        group.name,
                        this.getMembers(group.users),
                        <span>
                            <button onClick={(e) => {
                                this.props.courseMan.updateGroupStatus(group.id, CourseGroupStatus.approved);
                                this.props.navMan.refresh();
                            }}
                                className="btn btn-primary">
                                Approve
                        </button>
                            <button onClick={(e) => {
                                this.props.courseMan.updateGroupStatus(group.id, CourseGroupStatus.rejected);
                                this.props.navMan.refresh();
                            }} className="btn btn-danger">
                                Reject
                    </button>
                        </span>,
                    ]}
            />
        );
    }

    private getMembers(users: IUser[]): string {
        const names: string[] = [];
        for (const user of users) {
            names.push(user.firstname + " " + user.lastname);
        }
        return names.toString();
    }
}

export { CourseGroup };
