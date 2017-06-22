import * as React from "react";

interface IDynamicTableProps<T> {
    header: string[];
    data: T[];
    selector: (item: T) => string[];
}

class DynamicTable<T> extends React.Component<IDynamicTableProps<T>, undefined>{

    renderCells(values: string[]): JSX.Element[]{
        return values.map((v, i) => {
                return <td key={i}>{v}</td>
            });
    }

    renderRow(item: T, i: number): JSX.Element{
        return <tr key={i}>{ this.renderCells(this.props.selector(item)) }</tr>;
    }

    render(){
        let rows = this.props.data.map((v, i) => {
            return this.renderRow(v, i);
        });

        return <table className="table">
            <thead>
                <tr>{this.renderCells(this.props.header)}</tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    }
}

export {DynamicTable}