import * as React from "react";

import { CourseGroupStatus, ICourse, ICourseGroup } from "../../models";
import { Group } from "../../../proto/ag_pb";

interface IGroupPro {
    group: ICourseGroup;
    course: ICourse;
}
class GroupInfo extends React.Component<IGroupPro, any> {
    public render() {
        const groupMembers: JSX.Element[] = [];
        for (let i: number = 0; i < this.props.group.users.length; i++) {
            groupMembers.push(<li key={i} className="list-group-item">{this.props.group.users[i].getName()}</li>);
        }
        return (
            <div className="group-info">
                <h1>{this.props.course.name}</h1>
                <h3>{this.props.group.name} - <small>{this.getStatus()}</small></h3>
                <div className="group-members">
                    <ul className="list-group">
                        {groupMembers}
                    </ul>
                </div>
            </div>
        );
    }

    private getStatus(): string {
        switch (this.props.group.status) {
            case Group.GroupStatus.Approved:
                return "Appproved";
            case Group.GroupStatus.Pending:
                return "Pending";
            case Group.GroupStatus.Rejected:
                return "Rejected";
            default:
                return "N/A";
        }
    }
}
export { GroupInfo };
