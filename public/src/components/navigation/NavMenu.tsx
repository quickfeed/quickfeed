import * as React from "react";
import { ILink } from "../../managers/NavigationManager";
import { NavigationHelper } from "../../NavigationHelper";

interface INavMenuProps {
    links: ILink[];
    onClick?: (link: ILink) => void;
}

class NavMenu extends React.Component<INavMenuProps, undefined> {
    public render() {
        const items = this.props.links.map((v: ILink, i: number) => {
            let active = "";
            if (v.active) {
                active = "active";
            }
            return <li className={active} key={i}>
                <a href={"/" + v.uri}
                    onClick={(e) => NavigationHelper.handleClick(e, () => { this.handleClick(v); })}>
                    {v.name}
                </a>
            </li>;
        });
        return <ul className="nav nav-pills nav-stacked">
            {items}
        </ul>;
    }

    private handleClick(v: ILink) {
        if (this.props.onClick) {
            this.props.onClick(v);
        }
    }
}

export { INavMenuProps, NavMenu };
