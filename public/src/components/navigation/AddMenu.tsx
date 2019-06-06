
import * as React from "react";
import { ILink } from "../../managers";
import { User } from "../../../proto/ag_pb";
import { NavMenuDropdown } from "./NavMenuDropdown";

export interface IAddMenuProps {
    links?: ILink[];
    onClick?: (lin: ILink) => void;
    user: User | null;
}

interface IAddMenuState {
    loginOpen: boolean;
}

export class AddMenu extends React.Component<IAddMenuProps, IAddMenuState> {
    private lastCallback?: (e?: MouseEvent) => void;
    constructor(props: IAddMenuProps) {
        super(props);
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
        let isOpen = "";
        if (this.state.loginOpen) {
            isOpen = "open";
        }

        return <div className="navbar-right">
            <ul className="nav navbar-nav ">
                <li className={"dropdown " + isOpen}>
                    <a href="#"
                        style={{ padding: "15px" }}
                        onClick={(e) => { this.toggleMenu(e); }}
                        title="View Add options"
                        aria-haspopup="true"
                        aria-expanded="false" >
                        <span style={{ fontSize: "2em", verticalAlign: "middle" }}>+</span>
                        <span className="caret"></span>
                    </a>
                    <NavMenuDropdown links={links}
                        onClick={(link) => this.handleClick(link)}>
                    </NavMenuDropdown>
                </li>
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
        this.setState({ loginOpen: false });
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }
}
