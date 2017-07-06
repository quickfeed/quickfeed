import * as React from "react";
import { DynamicTable, Search } from "../../components";
import { NavigationManager, UserManager } from "../../managers";
import { IUser } from "../../models";

interface IUserViewerProps {
    users: IUser[];
    userMan?: UserManager;
    navMan?: NavigationManager;
    addSearchOption?: boolean;
}

interface IUserViewerState {
    users: IUser[];
}

export class UserView extends React.Component<IUserViewerProps, IUserViewerState> {

    public constructor(props: IUserViewerProps) {
        super(props);
        this.state = {
            users: props.users,
        };
    }

    public componentWillReceiveProps(nextProps: Readonly<IUserViewerProps>, nextContext: any): void {
        this.state = {
            users: nextProps.users,
        };
    }

    public render() {
        let searchForm: JSX.Element | null = null;
        if (this.props.addSearchOption) {
            const searchIcon: JSX.Element = <span className="input-group-addon">
                <i className="glyphicon glyphicon-search"></i>
            </span>;
            searchForm = <Search className="input-group"
                addonBefore={searchIcon}
                placeholder="Search for students"
                onChange={(query) => this.handleOnchange(query)}
            />;
        }
        return (
            <div>
                {searchForm}
                <DynamicTable
                    header={this.getTableHeading()}
                    data={this.state.users}
                    selector={(item: IUser) => this.getTableSelector(item)}
                />
            </div>);
    }

    private getTableHeading(): string[] {
        let heading: string[] = ["Name", "Email", "Student ID"];
        if (this.props.userMan) {
            heading = heading.concat("Action");
        }
        return heading;
    }

    private getTableSelector(user: IUser): Array<string | JSX.Element> {
        let selector: Array<string | JSX.Element> = [
            user.firstName + " " + user.lastName,
            <a href={"mailto:" + user.email}>{user.email}</a>,
            user.personId.toString(),
        ];
        if (this.props.userMan) {
            if (this.props.userMan.isAdmin(user)) {
                selector = selector.concat(
                    <button className="btn btn-danger"
                        onClick={() => this.handleAdminRoleClick(user)}
                        data-toggle="tooltip"
                        title="Demote from Admin">
                        Demote</button>);
            } else {
                selector = selector.concat(
                    <button className="btn btn-primary"
                        onClick={() => this.handleAdminRoleClick(user)}
                        data-toggle="tooltip"
                        title="Promote to Admin">
                        Promote</button>);
            }
        }
        return selector;
    }

    private async handleAdminRoleClick(user: IUser): Promise<boolean> {
        if (this.props.userMan && this.props.navMan) {
            const res = this.props.userMan.changeAdminRole(user);
            this.props.navMan.refresh();
            return res;
        }
        return false;
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IUser[] = [];
        this.props.users.forEach((user) => {
            if (user.firstName.toLowerCase().indexOf(query) !== -1
                || user.lastName.toLowerCase().indexOf(query) !== -1
                || user.email.toLowerCase().indexOf(query) !== -1
                || user.personId.toString().indexOf(query) !== -1
            ) {
                filteredData.push(user);
            }
        });

        this.setState({
            users: filteredData,
        });
    }
}
