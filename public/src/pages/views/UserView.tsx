import * as React from "react";
import { Enrollment, User } from "../../../proto/ag_pb";
import { BootstrapButton, BootstrapClass, DynamicTable, Search } from "../../components";
import { ILink, NavigationManager, UserManager } from "../../managers";

import { LiDropDownMenu } from "../../components/navigation/LiDropDownMenu";
import { generateLabRepoLink } from '../../helper';

interface IUserViewerProps {
    users: Enrollment[];
    isCourseList: boolean;
    userMan?: UserManager;
    navMan?: NavigationManager;
    courseURL: string;
    searchable?: boolean;
    actions?: ILink[];
    optionalActions?: (enrol: Enrollment) => ILink[];
    linkType?: ActionType;
    actionClick?: (enrollment: Enrollment, link: ILink) => void;
}

export enum ActionType {
    None,
    Menu,
    InRow,
}

interface IUserViewerState {
    enrollments: Enrollment[];
}

export class UserView extends React.Component<IUserViewerProps, IUserViewerState> {

    public constructor(props: IUserViewerProps) {
        super(props);
        this.state = {
            enrollments: props.users,
        };
    }

    public componentWillReceiveProps(nextProps: Readonly<IUserViewerProps>, nextContext: any): void {
        this.setState({
            enrollments: nextProps.users,
        });
    }

    public render() {
        return <div>
            {this.renderSearch()}
            <DynamicTable
                header={this.getTableHeading()}
                data={this.state.enrollments}
                classType={"table-grp"}
                selector={(item: Enrollment) => this.renderRow(item)}
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
            heading.push("Role");
        }
        return heading;
    }

    private renderRow(enr: Enrollment): (string | JSX.Element)[] {
        const selector: (string | JSX.Element)[] = [];
        const user = enr.getUser();
        if (!user) {
            return selector;
        }
        if (enr.getStatus() === Enrollment.UserStatus.TEACHER) {
            selector.push(
                <span className="text-muted">
                    <a href={this.gitLink(user.getLogin())} target="_blank">{user.getName()}</a>
                </span>);
        } else {
            selector.push(
                <a href={this.repoLink(user.getLogin())} target="_blank">{user.getName()}</a>);
        }
        selector.push(
            <a href={"mailto:" + enr.getUser()?.getEmail()}>{user?.getEmail()}</a>,
            enr.getUser()?.getStudentid() ?? "",
        );
        const temp = this.renderActions(enr);
        if (Array.isArray(temp) && temp.length > 0) {
            selector.push(<div className="btn-group action-btn">{temp}</div>);
        } else if (!Array.isArray(temp)) {
            selector.push(temp);
        }
        return selector;
    }

    private renderActions(enrol: Enrollment): JSX.Element[] | JSX.Element {
        const actionButtons: JSX.Element[] = [];
        const tempActions = this.getAllLinks(enrol);
        if (tempActions.length > 0) {
            switch (this.props.linkType) {
                case ActionType.Menu:
                    return this.renderDropdownMenu(enrol, tempActions);
                case ActionType.InRow:
                    actionButtons.push(...this.renderActionRow(enrol, tempActions));
                    break;
            }
        }
        return actionButtons;
    }

    private getAllLinks(enrol: Enrollment) {
        const tempActions: ILink[] = [];
        if (this.props.actions) {
            tempActions.push(...this.props.actions);
        }
        if (this.props.optionalActions) {
            tempActions.push(...this.props.optionalActions(enrol));
        }
        return tempActions;
    }

    private renderDropdownMenu(enrol: Enrollment, tempActions: ILink[]) {
        return <ul className="nav nav-pills">
            <LiDropDownMenu
                links={tempActions}
                onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(enrol, link); } }}>
                <span className="glyphicon glyphicon-option-vertical" />
            </LiDropDownMenu>
        </ul>;
    }

    private renderActionRow(enrol: Enrollment, tempActions: ILink[]) {
        return tempActions.map((v, i) => {
            let hoverText = "";
            if (v.uri === "teacher") {
                hoverText = "Promote to teacher";
            }
            if (v.uri === "demote") {
                hoverText = "Demote teacher";
            }

            return <BootstrapButton
                key={i}
                classType={v.extra ? v.extra as BootstrapClass : "default"}
                tooltip={hoverText}
                type={v.description}
                onClick={(link) => { if (this.props.actionClick) { this.props.actionClick(enrol, v); } }}
            >{v.name}
            </BootstrapButton>;
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: Enrollment[] = [];
        this.props.users.forEach((enr) => {
            const user = enr.toObject().user;
            if (user && (user.name.toLowerCase().indexOf(query) !== -1
                || user.email.toLowerCase().indexOf(query) !== -1
                || user.studentid.toString().indexOf(query) !== -1
                || user.login.toLowerCase().indexOf(query) !== -1
            )) {
                filteredData.push(enr);
            }
        });

        this.setState({
            enrollments: filteredData,
        });
    }

    private gitLink(user: string): string {
        return "https://github.com/" + user;
    }

    // return link to github account if there is no course information, otherwise return link to the student labs repo
    private repoLink(user: string): string {
        return this.props.isCourseList ? generateLabRepoLink(this.props.courseURL, user) : this.gitLink(user);
    }
}
