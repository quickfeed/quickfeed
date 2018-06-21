
import * as React from "react";
import { ILink } from "../../managers";
import { NavigationHelper } from "../../NavigationHelper";

import { IUser } from "../../models";
import { NavMenu } from "./NavMenu";
import { NavMenuDropdown } from "./NavMenuDropdown";

export interface ILiDropDownMenuProps {
    links?: ILink[];
    onClick?: (lin: ILink) => void;
}

interface ILiDropDownMenuState {
    loginOpen: boolean;
}

export class LiDropDownMenu extends React.Component<ILiDropDownMenuProps, ILiDropDownMenuState> {
    private lastCallback?: (e?: MouseEvent) => void;
    constructor(props: ILiDropDownMenuProps) {
        super(props);
        this.state = {
            loginOpen: false,
        };
    }

    public render(): JSX.Element {
        let links: ILink[] | undefined = this.props.links;
        if (!links) {
            links = [
                { name: "Missing links" },
            ];
        }
        let isOpen = "";
        if (this.state.loginOpen) {
            isOpen = "open";
        }

        return <li className={"dropdown " + isOpen}>
            <a href="#"
                onClick={(e) => { this.toggleMenu(e); }}
                title="View more"
                aria-haspopup="true"
                aria-expanded="false" >
                {this.props.children}
            </a>
            <NavMenuDropdown links={links}
                onClick={(link) => this.handleClick(link)}>
            </NavMenuDropdown>
        </li>;
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
        this.setState({ loginOpen: false });
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
