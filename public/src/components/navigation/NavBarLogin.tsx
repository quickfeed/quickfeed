
import * as React from "react";
import { ILink } from "../../managers";

import { IUser } from "../../models";
import { NavMenuDropdown } from "./NavMenuDropdown";

export interface INavBarLoginProps {
    links?: ILink[];
    onClick?: (lin: ILink) => void;
    user: IUser | null;
}

export interface INavBarLoginState {
    loginOpen: boolean;
}

export class NavBarLogin extends React.Component<INavBarLoginProps, INavBarLoginState> {
    private lastCallback?: (e?: MouseEvent) => void;
    constructor() {
        super();
        this.state = {
            loginOpen: false,
        };
    }
    public render(): JSX.Element {
        if (this.props.user) {
            const userMenuLinks: ILink[] = [
                { name: "Signed in as: " + this.props.user.name },
                { name: "#separator" },
                { name: "Your profile", uri: "/app/user" },
                { name: "Help", uri: "/app/help" },
                { name: "#separator" },
                { name: "Manage courses", uri: "app/admin/courses" },
                { name: "Manage users", uri: "app/admin/users" },
                { name: "#separator" },
                { name: "Sign out", uri: "app/login/logout" },
            ];
            let isOpen = "";
            if (this.state.loginOpen) {
                isOpen = "open";
            }
            return <div className="navbar-login">
                <ul className="nav navbar-nav navbar-right">
                    <li className={"dropdown " + isOpen}>
                        <a href="#"
                            title="View profile and more"
                            role="button"
                            onClick={(e) => { this.toggleMenu(e); }}
                            aria-haspopup="true"
                            aria-expanded="false">
                            <img className="img-rounded" src={this.props.user.avatarurl} width="20" height="20" />
                            <span className="caret"></span>
                        </a>
                        <NavMenuDropdown links={userMenuLinks}
                            onClick={(e) => { this.handleClick(e); }}>
                        </NavMenuDropdown>
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
                    <a onClick={(e) => { e.preventDefault(); this.handleClick(v); }}
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

    private toggleMenu(e: React.MouseEvent<HTMLAnchorElement>) {
        e.preventDefault();
        e.persist();
        if (this.lastCallback) {
            this.lastCallback();
            return;
        }
        this.lastCallback = (ev?: MouseEvent) => {
            console.log("callback");
            if (ev && ev.target === e.target) {
                return;
            }
            if (this.lastCallback) {
                window.removeEventListener("click", this.lastCallback as (ev: Event) => void);
                this.lastCallback = undefined;
            }
            this.setState({ loginOpen: false });
        };
        console.log("hello");
        window.addEventListener("click", this.lastCallback);
        console.log("opening");
        this.setState({ loginOpen: true });
    }

    private handleClick(link: ILink) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
