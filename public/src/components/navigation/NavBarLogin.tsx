
import * as React from "react";
import { ILink } from "../../managers";

import { IUser } from "../../models";

export interface INavBarLoginProps {
    links?: ILink[];
    onClick?: (lin: ILink) => void;
    user: IUser | null;
}

export class NavBarLogin extends React.Component<INavBarLoginProps, any> {
    public render(): JSX.Element {
        if (this.props.user) {
            return <div className="navbar-login pull-right">
                <ul className="nav navbar-nav navbar-right">
                    <li className="dropdown">
                        <a href="#"
                            title="View profile and more"
                            className="dropdown-toggle"
                            data-toggle="dropdown"
                            role="button"
                            aria-haspopup="true" aria-expanded="false">
                            <img className="img-rounded" src={this.props.user.avatarurl} width="20" height="20" />
                            <span className="caret"></span>
                        </a>
                        <ul className="dropdown-menu">
                            <li className="dropdown-header">
                                Signed in as &nbsp;&nbsp;
                                <strong>{this.props.user.name}</strong>
                            </li>
                            <li role="separator" className="divider"></li>
                            <li><a href="/app/user" onClick={(e) => {
                                this.handleClick(e, { name: "Profile", uri: "app/user" });
                            }}>Your Profile</a>
                            </li>
                            <li><a href="/app/help" onClick={(e) => {
                                this.handleClick(e, { name: "Help", uri: "app/help" });
                            }}> Help</a>
                            </li>
                            <li role="separator" className="divider"></li>
                            <li><a href="/app/admin/courses" onClick={(e) => {
                                this.handleClick(e, { name: "Manage courses", uri: "app/admin/courses" });
                            }}> Manage courses</a>
                            </li>
                            <li><a href="/app/admin/users" onClick={(e) => {
                                this.handleClick(e, { name: "Manage users", uri: "app/admin/users" });
                            }}> Manage users</a>
                            </li>
                            <li role="seperator" className="divider"></li>
                            <li><a href="app/login/logout"
                                onClick={(e) => {
                                    this.handleClick(e, { name: "Sign out", uri: "app/login/logout" });
                                }}>
                                Sign out</a>
                            </li>
                        </ul>
                    </li>
                </ul>
            </div >;
        }
        let links: ILink[] | undefined = this.props.links;
        if (!links) {
            links = [
                { name: "Missing links" },
            ];
        }

        const loginLinks = links.map((v: ILink, i: number) => {
            if (v.uri) {
                return <li key={i}>
                    <a onClick={(e) => this.handleClick(e, v)}
                        href={"/" + v.uri} title={v.name}>
                        <i className={"fa fa-2x fa-" + v.name.toLowerCase()} ></i>
                    </a>
                </li>;
            }
        });

        return <div className="navbar-login pull-right">
            <p className="navbar-text">Sign in with</p>
            <ul className="nav navbar-nav navbar-right social-login">
                {loginLinks}
            </ul>
        </div >;
    }

    private handleClick(e: any, link: ILink) {
        e.preventDefault();
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
