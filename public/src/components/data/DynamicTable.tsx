import * as React from "react";

interface IDynamicTableProps<T> {
    header: string[];
    data: T[];
    selector: (item: T) => string[];
    footer?: string[];
}

class DynamicTable<T> extends React.Component<IDynamicTableProps<T>, undefined> {
    public render() {
        const rows = this.props.data.map((v, i) => {
            return this.renderRow(v, i);
        });
        if (this.props.footer) {
            return (
                <table className="table">
                    <thead>
                        <tr>{this.renderCells(this.props.header)}</tr>
                    </thead>
                    <tbody>
                        {rows}
                    </tbody>
                    <tfoot>
                        <tr>{this.renderCells(this.props.footer)}</tr>
                    </tfoot>
                </table>
            );
        }
        return (
            <table className="table">
                <thead>
                    <tr>{this.renderCells(this.props.header)}</tr>
                </thead>
                <tbody>
                    {rows}
                </tbody>
            </table>
        );
    }

    private renderCells(values: string[]): JSX.Element[] {
        return values.map((v, i) => {
            return <td key={i}>{v}</td>;
        });
    }

    private renderRow(item: T, i: number): JSX.Element {
        return <tr key={i}>{this.renderCells(this.props.selector(item))}</tr>;

    }
}

export { DynamicTable };
