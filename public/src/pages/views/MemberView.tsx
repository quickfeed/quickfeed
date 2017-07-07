import * as React from "react";
import { CourseManager, NavigationManager, UserManager } from "../../managers";
import { CourseUserState, ICourse, ICourseUserLink, IUser } from "../../models";

import { DynamicTable } from "../../components";
import { UserView } from "./UserView";

interface IUserViewerProps {
    users: IUser[];
    navMan: NavigationManager;
    courseMan: CourseManager;
    acceptedUsers: IUser[];
    pendingUsers: Array<{ ele1: ICourseUserLink, ele2: IUser }>;
    course: ICourse;
}

export interface IUserViewerState {
    users: IUser[];
}

export class MemberView extends React.Component<IUserViewerProps, IUserViewerState> {
    constructor(props: any) {
        super(props);
    }

    public render() {

        let condPending;
        if (this.props.pendingUsers.length > 0) {
            condPending = <div><h3>Pending users</h3>{this.createPendingTable(this.props.pendingUsers)}</div>;
        }
        const userView = <div><h3>Registered users</h3><UserView users={this.props.acceptedUsers}></UserView></div>;
        return <div>
            <h1>{this.props.course.name}</h1>
            {userView}
            {condPending}
        </div>;
    }

    public createPendingTable(pendingUsers: Array<{ ele1: ICourseUserLink, ele2: IUser }>): JSX.Element {
        return <DynamicTable
            data={pendingUsers}
            header={["Name", "Email", "Student ID", "Action"]}
            selector={
                (ele: { ele1: ICourseUserLink, ele2: IUser }) => [
                    ele.ele2.firstName + " " + ele.ele2.lastName,
                    <a href={"mailto:" + ele.ele2.email}>{ele.ele2.email}</a>,
                    ele.ele2.personId.toString(),
                    <span>
                        <button onClick={(e) => {
                            this.props.courseMan.changeUserState(ele.ele1, CourseUserState.student);
                            this.props.navMan.refresh();
                        }}
                            className="btn btn-primary">
                            Accept
                    </button>
                        <button onClick={(e) => {
                            this.props.courseMan.changeUserState(ele.ele1, CourseUserState.rejected);
                            this.props.navMan.refresh();
                        }} className="btn btn-danger">
                            Reject
                    </button>
                    </span>,
                ]}
        >
        </DynamicTable>;
    }
}

export { UserView };
