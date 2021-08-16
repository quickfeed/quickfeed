import * as React from "react";
import { ILink } from "../../managers/NavigationManager";
import { NavigationHelper } from "../../NavigationHelper";

interface INavMenuDropdownProps {
    links: ILink[];
    onClick?: (link: ILink) => void;
}

class NavMenuDropdown extends React.Component<INavMenuDropdownProps, {}> {
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
            } else if (v.name === "#separator") {
                return <li key={i} className="divider" role="separator" />;
            } else {
                return <li key={i} className={"dropdown-header " + active}>
                    {v.name}
                </li>;
            }
        });
        return <ul className="dropdown-menu">
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

export { INavMenuDropdownProps, NavMenuDropdown };
