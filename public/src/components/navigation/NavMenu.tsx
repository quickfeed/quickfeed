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
            if (v.uri) {
                return <li key={i} className={active}>
                    <a onClick={(e) => this.handleClick(e, v)}
                        href={"/" + v.uri}>{v.name}</a>
                </li>;
            } else {
                return <li key={i} className={active}>
                    <span className="header">{v.name}</span>
                </li>;
            }
        });
        return <ul className="nav nav-list">
            {items}
        </ul>;
    }

    private handleClick(e: React.MouseEvent<HTMLAnchorElement>, link: ILink) {
        NavigationHelper.handleClick(e, () => {
            if (this.props.onClick) {
                this.props.onClick(link);
            }
        });
    }
}

export { INavMenuProps, NavMenu };
