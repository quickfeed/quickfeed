import * as React from "react";
import { ILink } from "../../managers/NavigationManager";

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
                <a href={"/" + v.uri} onClick={(e) => this.handleClick(e, v)}>
                    {v.name}
                </a>
            </li>;
        });
        return <ul className="nav nav-pills nav-stacked">
            {items}
        </ul>;
    }

    private handleClick(e: React.MouseEvent<HTMLAnchorElement>, v: ILink) {
        e.preventDefault();
        if (this.props.onClick) {
            this.props.onClick(v);
        }
    }
}

export { INavMenuProps, NavMenu };
