import * as React from "React";

interface INavMenuFormatableProps<T> {
    items: T[];
    formater?: (item: T) => string;
    onClick?: (item: T) => void;
}

class NavMenuFormatable<T> extends React.Component<INavMenuFormatableProps<T>, undefined> {
    renderObj(item: T): string{
        if (this.props.formater){
            return this.props.formater(item);
        }
        return item.toString();
    }

    handleItemClick(item: T): void{
        if (this.props.onClick){
            this.props.onClick(item);
        }
    }

    render(){
        const items = this.props.items.map((v, i) => {
            return <li key={i}><a href="#" onClick={() => { this.handleItemClick(v) }}>{this.renderObj(v)}</a></li>
        })
        return <ul className="nav nav-pills nav-stacked">
            {items}
        </ul>
    }
}

export {INavMenuFormatableProps, NavMenuFormatable}