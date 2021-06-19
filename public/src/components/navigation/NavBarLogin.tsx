
import * as React from "react";
import { User } from "../../../proto/ag/ag_pb";
import { ILink } from "../../managers";
import { NavMenuDropdown } from "./NavMenuDropdown";

export interface INavBarLoginProps {
    loginLinks?: ILink[];
    userLinks?: ILink[];
    onClick?: (lin: ILink) => void;
    user: User | null;
}

export interface INavBarLoginState {
    loginOpen: boolean;
}

export class NavBarLogin extends React.Component<INavBarLoginProps, INavBarLoginState> {
    private lastCallback?: (e?: MouseEvent) => void;
    constructor(props: INavBarLoginProps) {
        super(props);
        this.state = {
            loginOpen: false,
        };
    }
    public render(): JSX.Element {
        if (this.props.user && this.props.userLinks) {
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
                            <img className="img-rounded" src={this.props.user.getAvatarurl()} width="20" height="20" />
                            <span className="caret"></span>
                        </a>
                        <NavMenuDropdown links={this.props.userLinks}
                            onClick={(e) => { this.handleClick(e); }}>
                        </NavMenuDropdown>
                    </li>
                </ul>
            </div >;
        }
        let links: ILink[] | undefined = this.props.loginLinks;
        if (!links) {
            links = [
                { name: "Missing links" },
            ];
        }

        const loginLinks = links.map((v: ILink, i: number) => {
            if (v.uri) {
                return <li key={i}>
                    <a onClick={(e) => { e.preventDefault(); this.handleClick(v); }}
                        href={"/" + v.uri} title={this.stringifyLink(v.name)}>
                        <i className={"fa fa-2x fa-" + this.stringifyLink(v.name).toLowerCase()} ></i>
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
            if (ev && ev.target === e.target) {
                return;
            }
            if (this.lastCallback) {
                window.removeEventListener("click", this.lastCallback as (ev: Event) => void);
                this.lastCallback = undefined;
            }
            this.setState({ loginOpen: false });
        };
        window.addEventListener("click", this.lastCallback);
        this.setState({ loginOpen: true });
    }

    private handleClick(link: ILink) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }

    // link name can come as a JSX.Element, for example in a case of a button
    // with a glyphicon. In such a case, treat the name as if it was an empty string
    private stringifyLink(linkName?: string | JSX.Element): string {
        if (linkName && (linkName instanceof Element || typeof(linkName) === "string")) {
            return linkName.toString();
        }
        return "";
    }
}
