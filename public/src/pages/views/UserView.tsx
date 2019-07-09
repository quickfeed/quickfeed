import * as React from "react";
import { Enrollment} from "../../../proto/ag_pb";
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink, NavigationManager, UserManager } from "../../managers";
import { IUserRelation } from "../../models";

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
        const state = {users: nextProps.users};
        this.setState(state);
    }

    public render() {
        return <div>
            {this.renderSearch()}
            <DynamicTable
                header={this.getTableHeading()}
                data={this.state.users}
                selector={(item: IUserRelation) => this.renderRow(item)}
            />
        </div>;
    }

    private renderSearch() {
        if (this.props.searchable) {
            return <Search className="input-group"
                placeholder="Search for students"
                onChange={(query) => this.handleOnchange(query)}
            />;
        }
        return null;
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
        if (user.link.state === Enrollment.UserStatus.TEACHER) {
            selector.push(
                <span className="text-muted">
                    <a href={"https://github.com/" + user.user.getLogin()} target="_blank">{user.user.getName()}</a>
                </span>
            );
        } else {
            selector.push(
                <a href={"https://github.com/" + user.user.getLogin()} target="_blank">{user.user.getName()}</a>
                );
        }
        selector.push(
            <a href={"mailto:" + user.user.getEmail()}>{user.user.getEmail()}</a>,
            user.user.getStudentid().toString(),
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
        const tempActions = this.getAllLinks(user);
        if (tempActions.length > 0) {
            switch (this.props.linkType) {
                case ActionType.Menu:
                    return this.renderDropdownMenu(user, tempActions);
                case ActionType.InRow:
                    actionButtons.push(...this.renderActionRow(user, tempActions));
                    break;
            }
        }
        return actionButtons;
    }

    private getAllLinks(user: IUserRelation) {
        const tempActions: ILink[] = [];
        if (this.props.actions) {
            tempActions.push(...this.props.actions);
        }
        if (this.props.optionalActions) {
            tempActions.push(...this.props.optionalActions(user));
        }
        return tempActions;
    }

    private renderDropdownMenu(user: IUserRelation, tempActions: ILink[]) {
        return <ul className="nav nav-pills">
            <LiDropDownMenu
                links={tempActions}
                onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(user, link); } }}>
                <span className="glyphicon glyphicon-option-vertical" />
            </LiDropDownMenu>
        </ul>;
    }

    private renderActionRow(user: IUserRelation, tempActions: ILink[]) {
        return tempActions.map((v, i) => {
            return <BootstrapButton
                key={i}
                classType={v.extra ? v.extra as BootstrapClass : "default"}
                onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(user, v); } }}
            >{v.name}
            </BootstrapButton>;
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IUserRelation[] = [];
        this.props.users.forEach((user) => {
            if (user.user.getName().toLowerCase().indexOf(query) !== -1
                || user.user.getEmail().toLowerCase().indexOf(query) !== -1
                || user.user.getStudentid().toString().indexOf(query) !== -1
            ) {
                filteredData.push(user);
            }
        });

        this.setState({
            users: filteredData,
        });
    }
}
