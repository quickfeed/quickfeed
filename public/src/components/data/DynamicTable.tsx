import * as React from "react";

interface IDynamicTableProps<T> {
    header: string[];
    data: T[];
    selector: (item: T) => string[];
    footer?: string[];
    onRowClick?: (link: string) => void;
    row_links?: { [lab: string]: string };
    link_key_identifier?: string;
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

    private renderCells(values: string[], th: boolean = false): JSX.Element[] {
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
                onClick={(e) => this.handleRowClick(e, item)}>
                {this.renderCells(this.props.selector(item))}
            </tr>
        );

    }

    private tableWithFooter(rows: any, footer: any): JSX.Element {
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

    private tableWithNoFooter(rows: any) {
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

    private handleRowClick(e: React.MouseEvent<HTMLTableRowElement>, item: any) {
        e.preventDefault();
        if (this.props.onRowClick && this.props.row_links && this.props.link_key_identifier) {
            const identifier = this.props.link_key_identifier;
            this.props.onRowClick(this.props.row_links[item[identifier]]);
        }
    }
}

export {DynamicTable};
