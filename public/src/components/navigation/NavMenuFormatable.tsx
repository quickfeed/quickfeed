import * as React from "react";

interface INavMenuFormatableProps<T> {
    items: T[];
    formater?: (item: T) => string;
    onClick?: (item: T) => void;
}

class NavMenuFormatable<T> extends React.Component<INavMenuFormatableProps<T>, undefined> {
    public render() {
        const items = this.props.items.map((v, i) => {
            return <li key={i}><a href="#" onClick={() => { this.handleItemClick(v); }}>{this.renderObj(v)}</a></li>;
        });
        return <ul className="nav nav-pills nav-stacked">
            {items}
        </ul>;
    }

    private renderObj(item: T): string {
        if (this.props.formater) {
            return this.props.formater(item);
        }
        return item.toString();
    }

    private handleItemClick(item: T): void {
        if (this.props.onClick) {
            this.props.onClick(item);
        }
    }
}

export { INavMenuFormatableProps, NavMenuFormatable };
