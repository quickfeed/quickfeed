import * as React from "react";
import { NavigationHelper } from "../../NavigationHelper";

interface INavHeaderBarProps {
    brandName: string;
    id: string;
    brandClick: () => void;
}

class NavHeaderBar extends React.Component<INavHeaderBarProps, {}> {
    public componentDidMount() {
        const temp = this.refs.button as HTMLElement;
        temp.setAttribute("data-toggle", "collapse");
        temp.setAttribute("data-target", "#" + this.props.id);
        temp.setAttribute("aria-expanded", "false");
    }

    public render() {
        return <div className="navbar-header">
            <button ref="button" type="button" className="navbar-toggle collapsed" >
                <span className="sr-only">Toggle navigation</span>
                <span className="icon-bar"></span>
                <span className="icon-bar"></span>
                <span className="icon-bar"></span>
            </button>
            <a className="navbar-brand" onClick={(e) => {
                NavigationHelper.handleClick(e, () => {
                    this.props.brandClick();
                });
            }} href=";/">
                {this.props.brandName}
            </a>
        </div>;
    }
}

export { NavHeaderBar, INavHeaderBarProps };
