import * as React from "react";
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink, NavigationManager, UserManager } from "../../managers";
import { CourseUserState, IUser, IUserRelation } from "../../models";

import { LiDropDownMenu } from "../../components/navigation/LiDropDownMenu";

interface IUserViewerProps {
    users: IUserRelation[];
    userMan?: UserManager;
    navMan?: NavigationManager;
    searchable?: boolean;
    actions?: ILink[];
    optionalActions?: (user: IUserRelation) => ILink[];
    linkType?: ActionType;
    actionClick?: (user: IUserRelation, link: ILink) => void;
}

export enum ActionType {
    None,
    Menu,
    InRow,
}

interface IUserViewerState {
    users: IUserRelation[];
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
        if (this.props.searchable) {
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
                    selector={(item: IUserRelation) => this.renderRow(item)}
                />
            </div>);
    }

    private getTableHeading(): string[] {
        const heading: string[] = ["Name", "Email", "Student ID"];
        if (this.props.userMan || this.props.actions) {
            heading.push("Options");
        }
        return heading;
    }

    private renderRow(user: IUserRelation): Array<string | JSX.Element> {
        const selector: Array<string | JSX.Element> = [];
        if (user.link.state === CourseUserState.teacher) {
            selector.push(<span className="text-muted">{user.user.name}</span>);
        } else {
            selector.push(user.user.name);
        }
        selector.push(
            <a href={"mailto:" + user.user.email}>{user.user.email}</a>,
            user.user.studentid.toString(),
        );
        const temp = this.renderActions(user);
        if (Array.isArray(temp) && temp.length > 0) {
            selector.push(<div>{temp}</div>);
        } else if (!Array.isArray(temp)) {
            selector.push(temp);
        }
        return selector;
    }

    private renderActions(user: IUserRelation): JSX.Element[] | JSX.Element {
        const actionButtons: JSX.Element[] = [];
        const tempActions: ILink[] = [];
        if (this.props.actions) {
            tempActions.push(...this.props.actions);
        }
        if (this.props.optionalActions) {
            tempActions.push(...this.props.optionalActions(user));
        }
        if (tempActions.length > 0) {
            if (this.props.linkType === ActionType.Menu) {
                return <ul className="nav nav-pills">
                    <LiDropDownMenu
                        links={tempActions}
                        onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(user, link); } }}>
                        <span className="glyphicon glyphicon-option-vertical" />
                    </LiDropDownMenu>
                </ul>;
            } else if (this.props.linkType === ActionType.InRow) {
                actionButtons.push(...tempActions.map((v, i) => {
                    return <BootstrapButton
                        key={i}
                        classType={v.extra ? v.extra as BootstrapClass : "default"}
                        onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(user, v); } }}
                    >{v.name}
                    </BootstrapButton>;
                }));
            }
        }

        if (this.props.userMan) {
            if (this.props.userMan.isAdmin(user.user)) {
                actionButtons.push(<button className="btn btn-danger"
                    onClick={() => this.handleAdminRoleClick(user.user)}
                    data-toggle="tooltip"
                    title="Demote from Admin">
                    Demote</button>);
            } else {
                actionButtons.push(<button className="btn btn-primary"
                    onClick={() => this.handleAdminRoleClick(user.user)}
                    data-toggle="tooltip"
                    title="Promote to Admin">
                    Promote</button>);
            }
            return actionButtons;
        }
        return actionButtons;
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
        const filteredData: IUserRelation[] = [];
        this.props.users.forEach((user) => {
            if (user.user.name.toLowerCase().indexOf(query) !== -1
                || user.user.email.toLowerCase().indexOf(query) !== -1
                || user.user.studentid.toString().indexOf(query) !== -1
            ) {
                filteredData.push(user);
            }
        });

        this.setState({
            users: filteredData,
        });
    }
}
