
import * as React from "React";
import { ILink } from "../../managers/NavigationManager";

interface INavDropdownProps {
    items: ILink[];
    selectedIndex: number;
    itemClick: (item: ILink, index: number) => void;
}

interface INavDropdownState {
    isOpen: boolean;
}

class NavDropdown extends React.Component<INavDropdownProps, INavDropdownState> {
    constructor() {
        super();
        this.state = {
            isOpen: false,
        };
    }

    public render() {
        const children = this.props.items.map((item, index) => {
            return <li key={index}>
                <a href={"/" + item.uri} onClick={(e) => {
                    e.preventDefault();
                    this.toggleOpen();
                    this.props.itemClick(item, index);
                }}>
                    {item.name}
                </a>
            </li>;
        });

        return <div className={this.getButtonClass()}>
            <button
                className="btn btn-default dropdown-toggle"
                type="button"
                // id="dropdownMenu1" data-toggle="dropdown" aria-haspopup="true" aria-expanded="true"
                onClick={() => this.toggleOpen()}
            >
                {this.renderActive()}
                <span className="caret"></span>
            </button >
            <ul className="dropdown-menu">
                {children}
            </ul>
        </div >;
    }

    private getButtonClass(): string {
        if (this.state.isOpen) {
            return "button open";
        } else {
            return "button";
        }
    }

    private toggleOpen(): void {
        const newState = !this.state.isOpen;
        this.setState({ isOpen: newState });
    }

    private renderActive(): string {
        console.log(this.props);
        if (this.props.items.length === 0) {
            return "";
        }
        let curIndex = this.props.selectedIndex;
        if (curIndex >= this.props.items.length) {
            curIndex = 0;
        }
        return this.props.items[curIndex].name;
    }
}

export { NavDropdown, INavDropdownProps };
