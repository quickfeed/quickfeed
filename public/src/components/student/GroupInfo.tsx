import * as React from "react";

import { Course, Group } from "../../../proto/ag_pb";

interface IGroupPro {
    group: Group;
    course: Course;
}
class GroupInfo extends React.Component<IGroupPro, any> {
    public render() {
        const groupMembers: JSX.Element[] = [];
        const users = this.props.group.getUsersList();
        for (let i: number = 0; i < users.length; i++) {
            groupMembers.push(<li key={i} className="list-group-item">{users[i].getName()}</li>);
        }
        return (
            <div className="group-info">
                <h1>{this.props.course.getName()}</h1>
                <h3>{this.props.group.getName()} - <small>{this.getStatus()}</small></h3>
                <div className="group-members">
                    <ul className="list-group">
                        {groupMembers}
                    </ul>
                </div>
            </div>
        );
    }

    private getStatus(): string {
        switch (this.props.group.getStatus()) {
            case Group.GroupStatus.APPROVED:
                return "Appproved";
            case Group.GroupStatus.PENDING:
                return "Pending";
            case Group.GroupStatus.REJECTED:
                return "Rejected";
            default:
                return "N/A";
        }
    }
}
export { GroupInfo };
