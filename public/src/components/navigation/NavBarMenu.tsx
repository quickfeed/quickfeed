
import * as React from "react";
import { ILink } from "../../managers";
import { NavigationHelper } from "../../NavigationHelper";

export interface INavBarMenuProps {
    links: ILink[];
    onClick?: (lin: ILink) => void;
}

export class NavBarMenu extends React.Component<INavBarMenuProps, {}> {
    public render(): JSX.Element {
        const items = this.props.links.map((link, i) => {
            let active = "";
            if (link.active) {
                active = "active";
            }
            return <li className={active} key={i}>
                <a href={"/" + link.uri} onClick={(e) => {
                    NavigationHelper.handleClick(e, () => {
                        this.handleClick(link);
                    });
                }}>
                    {link.name}
                </a>
            </li>;
        });

        return <ul className="nav navbar-nav">
            {items}
        </ul>;
    }

    private handleClick(link: ILink) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
