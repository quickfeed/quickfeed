import * as React from "react";

import { ILink } from "../../managers/NavigationManager";
import { IUser } from "../../models";
import { NavHeaderBar } from "./NavHeaderBar";

import { NavigationHelper } from "../../NavigationHelper";
import { NavMenu } from "./NavMenu";

interface INavBarProps {
    id: string;
    isFluid: boolean;
    isInverse: boolean;
    brandName: string;
    onClick?: (lin: ILink) => void;
}

class NavBar extends React.Component<INavBarProps, {}> {

    public render() {
        return <nav className={this.renderNavBarClass()}>
            <div className={this.renderIsFluid()}>
                <NavHeaderBar
                    id={this.props.id}
                    brandName={this.props.brandName}
                    brandClick={() => this.handleClick({ name: "Home", uri: "/" })}>
                </NavHeaderBar>

                <div className="collapse navbar-collapse" id={this.props.id}>
                    {this.props.children}
                </div>
            </div>
        </nav>;
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
