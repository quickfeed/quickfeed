
import * as React from "react";
import { ILink } from "../../managers";
import { NavigationHelper } from "../../NavigationHelper";

import { IUser } from "../../models";
import { NavMenu } from "./NavMenu";

export interface IAddMenuProps {
    links?: ILink[];
    onClick?: (lin: ILink) => void;
    user: IUser | null;
}

interface IAddMenuState {
    loginOpen: boolean;
}

export class AddMenu extends React.Component<IAddMenuProps, IAddMenuState> {

    constructor() {
        super();
        this.state = {
            loginOpen: false,
        };
    }

    public render(): JSX.Element {
        if (!this.props.user) {
            return <div></div>;
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
            <ul className="nav navbar-nav ">
                <li>
                    <a style={{ padding: "13px 15px 12px 15px" }}
                        href="#" onClick={(e) => { e.preventDefault(); this.toggleMenu(); }}
                        className="">
                        <span style={{ fontSize: "2em", verticalAlign: "middle" }}>+</span>
                        <span style={{ fontSize: "0.5em", verticalAlign: "sub" }}>&#9660;</span>
                    </a>
                    <div className={"nav-box " + isHidden}>
                        <NavMenu links={links}
                            onClick={(link) => this.handleClick(link)}>
                        </NavMenu>
                    </div>
                </li>
            </ul>

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
