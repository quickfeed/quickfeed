import * as React from "react";

import {NavHeaderBar} from "./NavHeaderBar"
import { ILink } from "../../managers/NavigationManager";
import { IUser } from "../../models";

interface INavBarProps{
    id: string;
    links: ILink[];
    isFluid: boolean;
    isInverse: boolean;
    brandName: string;
    onClick?: (link:ILink) => void;
    user: IUser | null;
}

class NavBar extends React.Component<INavBarProps, undefined> {

    private renderIsFluid(){
        let name = "container"
        if (this.props.isFluid){
            name += "-fluid";
        }
        return name;
    }

    private renderNavBarClass(){
        let name = "navbar navbar-absolute-top";
        if (this.props.isInverse){
            name += " navbar-inverse";
        }
        else 
        {
            name += " navbar-default";
        }
        return name;
    }

    private handleClick(link: ILink){
        if (this.props.onClick){
            this.props.onClick(link);
        }
    }

    renderUser(user: IUser | null): string{
        if (user){
            return "Hello " + user.firstName;
        }
        return "Not logged in";
    }

    render(){
        let items = this.props.links.map((v, i) => {
            let active = "";
            if(v.active){
                active = "active";
            }
            return <li className={active} key={i}><a href={"/" + v.uri}  onClick={(e) => { e.preventDefault(); this.handleClick(v); }}>{v.name}</a></li>
        });

        return <nav className={this.renderNavBarClass()}>
            <div className={this.renderIsFluid()}>
                <NavHeaderBar id={this.props.id} brandName={this.props.brandName} brandClick={() => this.handleClick({name:"Home", uri: "/"})}></NavHeaderBar>

                <div className="collapse navbar-collapse" id={this.props.id}>
                    <ul className="nav navbar-nav">
                        {items}
                    </ul>
                    <p className="navbar-text navbar-right">
                        { this.renderUser(this.props.user) }
                    </p>
                </div>
            </div>
        </nav>
    }
}

export {NavBar, INavBarProps};