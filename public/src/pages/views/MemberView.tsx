import * as React from "react";
import { CourseManager, NavigationManager, UserManager } from "../../managers";
import { CourseUserState, ICourse, ICourseUserLink, IUser, IUserRelation } from "../../models";

import { DynamicTable } from "../../components";
import { UserView } from "./UserView";

interface IUserViewerProps {
    navMan: NavigationManager;
    courseMan: CourseManager;
    acceptedUsers: IUserRelation[];
    pendingUsers: IUserRelation[];
    course: ICourse;
}

export class MemberView extends React.Component<IUserViewerProps, {}> {
    public render() {
        let condPending;
        if (this.props.pendingUsers.length > 0) {
            condPending = <div><h3>Pending users</h3>{this.createPendingTable(this.props.pendingUsers)}</div>;
        }
        const userView = <div>
            <h3>Registered users</h3>
            <UserView users={this.props.acceptedUsers.map((userRel) => userRel.user)}>

            </UserView>
        </div>;
        return <div>
            <h1>{this.props.course.name}</h1>
            {userView}
            {condPending}
        </div>;
    }

    public createPendingTable(pendingUsers: IUserRelation[]): JSX.Element {
        return <DynamicTable
            data={pendingUsers}
            header={["Name", "Email", "Student ID", "Action"]}
            selector={
                (userRel: IUserRelation) => [
                    userRel.user.firstname + " " + userRel.user.lastname,
                    <a href={"mailto:" + userRel.user.email}>{userRel.user.email}</a>,
                    userRel.user.studentnr.toString(),
                    <span>
                        <button onClick={(e) => {
                            this.props.courseMan.changeUserState(userRel.link, CourseUserState.student);
                            this.props.navMan.refresh();
                        }}
                            className="btn btn-primary">
                            Accept
                        </button>
                        <button onClick={(e) => {
                            this.props.courseMan.changeUserState(userRel.link, CourseUserState.rejected);
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
