import * as React from "react";

interface IDynamicTableProps<T> {
    header: string[];
    data: T[];
    selector: (item: T) => Array<string | JSX.Element>;
    footer?: string[];
    onRowClick?: (link: T) => void;
}

class DynamicTable<T> extends React.Component<IDynamicTableProps<T>, undefined> {

    public render() {
        const rows = this.props.data.map((v, i) => {
            return this.renderRow(v, i);
        });
        if (this.props.footer) {
            return this.tableWithFooter(rows, this.props.footer);
        }
        return this.tableWithNoFooter(rows);
    }

    private renderCells(values: Array<string | JSX.Element>, th: boolean = false): JSX.Element[] {
        return values.map((v, i) => {
            if (th) {
                return <th key={i}>{v}</th>;
            }
            return <td key={i}>{v}</td>;
        });
    }

    private renderRow(item: T, i: number): JSX.Element {
        return (
            <tr key={i}
                onClick={(e) => this.handleRowClick(item)}>
                {this.renderCells(this.props.selector(item))}
            </tr>
        );

    }

    private tableWithFooter(rows: JSX.Element[], footer: string[]): JSX.Element {
        return (
            <table className={this.props.onRowClick ? "table table-hover" : "table"}>
                <thead>
                    <tr>{this.renderCells(this.props.header, true)}</tr>
                </thead>
                <tbody>
                    {rows}
                </tbody>
                <tfoot>
                    <tr>{this.renderCells(footer)}</tr>
                </tfoot>
            </table>
        );
    }

    private tableWithNoFooter(rows: JSX.Element[]) {
        return (
            <table className={this.props.onRowClick ? "table table-hover" : "table"}>
                <thead>
                    <tr>{this.renderCells(this.props.header, true)}</tr>
                </thead>
                <tbody>
                    {rows}
                </tbody>
            </table>
        );
    }

    private handleRowClick(item: T) {
        if (this.props.onRowClick) {
            this.props.onRowClick(item);
        }
    }
}

export { DynamicTable };
