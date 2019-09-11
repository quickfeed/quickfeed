import * as React from "react";

import { ILink } from "../../managers/NavigationManager";
import { NavHeaderBar } from "./NavHeaderBar";

interface INavBarProps {
    id: string;
    isFluid: boolean;
    isInverse: boolean;
    brandName: string;
    onClick?: (lin: ILink) => void;
}

interface INavBarState {
    collapsed: boolean;
}

class NavBar extends React.Component<INavBarProps, {}> {

    public state: INavBarState = {
        collapsed: true,
    };

    public render() {
        return <nav className={this.renderNavBarClass()}>
            <div className={this.renderIsFluid()}>
                <NavHeaderBar
                    id={this.props.id}
                    brandName={this.props.brandName}
                    isCollapsed={this.state.collapsed}
                    brandClick={() => this.handleClick({ name: "Home", uri: "/" })}
                    toggleNavbar={this.toggleNavbar}>
                </NavHeaderBar>

                <div
                    className={`collapse navbar-collapse ${this.state.collapsed ? "" : "show"}`}
                    id={this.props.id}
                >
                    {this.props.children}
                </div>
            </div>
        </nav>;
    }

    private toggleNavbar = () => {
        this.setState((state: INavBarState) => {
            return {
                collapsed: !state.collapsed,
            };
        });
    }

    private handleClick(link: ILink) {
        if (this.props.onClick) {
            this.props.onClick(link);
        }
    }

    private renderIsFluid() {
        let name = "container";
        if (this.props.isFluid) {
            name += "-fluid";
        }
        return name;
    }

    private renderNavBarClass() {
        let name = "navbar navbar-absolute-top spacefix";
        if (this.props.isInverse) {
            name += " navbar-inverse spacefix";
        } else {
            name += " navbar-default spacefix";
        }
        return name;
    }
}

export { NavBar, INavBarProps };
