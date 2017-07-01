
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

    constructor() {
        super();
        this.state = {
            loginOpen: false,
        };
    }

    public render(): JSX.Element {
        if (this.props.user) {
            return <div className="navbar-right">
                <button className="btn btn-primary navbar-btn"
                    onClick={() => { this.handleClick({ name: "Logout", uri: "app/login/logout" }); }}>
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
        let isHidden = "hidden";
        if (this.state.loginOpen) {
            isHidden = "";
        }

        return <div className="navbar-right">
            <button onClick={() => this.toggleMenu()}
                className="btn btn-primary navbar-btn">
                Login
            </button>
            <div className={"nav-box " + isHidden}>
                <NavMenu links={links}
                    onClick={(link) => this.handleClick(link)}>
                </NavMenu>
            </div>
        </div >;
    }

    private toggleMenu() {
        this.setState({ loginOpen: !this.state.loginOpen });
    }

    private handleClick(link: ILink) {
        this.setState({ loginOpen: false });
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
