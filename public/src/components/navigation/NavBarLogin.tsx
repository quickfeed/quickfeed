
import * as React from "react";
import { ILink } from "../../managers";
import { NavigationHelper } from "../../NavigationHelper";

import { IUser } from "../../models";
import { NavMenu } from "./NavMenu";

export interface INavBarLoginProps {
    links?: ILink[];
    onClick?: (lin: ILink) => void;
    user: IUser | null;
}

interface INavBarLoginState {
    loginOpen: boolean;
}

export class NavBarLogin extends React.Component<INavBarLoginProps, INavBarLoginState> {

    public render(): JSX.Element {
        if (this.props.user) {
            return <div className="navbar-right">
                <button className="btn btn-primary navbar-btn"
                    onClick={() => { this.handleClick({ name: "Sign out", uri: "app/login/logout" }); }}>
                    Log out
                </button>
            </div>;
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
                    <a onClick={() => this.handleClick(v) }
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

    private handleClick(link: ILink) {
        this.setState({ loginOpen: false });
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
