import * as React from "react";

interface IDynamicTableProps<T> {
    header: string[];
    footer?: string[];
    data: T[];
    selector: (item: T) => Array<string | JSX.Element | ICellElement>;
    onRowClick?: (link: T) => void;
}

export interface ICellElement {
    value: string | JSX.Element;
    className?: string;
}

function isICellElement(obj: any): obj is ICellElement {
    return obj.value;
}

export class DynamicTable<T> extends React.Component<IDynamicTableProps<T>, {}> {

    public render() {
        const footer = this.props.footer;
        const rows = this.props.data.map((v, i) => {
            return this.renderRow(v, i);
        });
        const tableFooter = footer ? <tfoot><tr>{this.renderCells(footer)}</tr></tfoot> : null;

        return (
            <table className={this.props.onRowClick ? "table table-hover" : "table"}>
                <thead>
                    <tr>{this.renderCells(this.props.header, true)}</tr>
                </thead>
                <tbody>
                    {rows}
                </tbody>
                {tableFooter}
            </table>
        );
    }

    private renderCells(values: Array<string | JSX.Element | ICellElement>, th: boolean = false): JSX.Element[] {
        return values.map((v, i) => {
            if (th) {
                if (isICellElement(v)) {
                    return <th key={i} className={v.className ? v.className : ""}>{v.value}</th>;
                } else {
                    return <th key={i}>{v}</th>;
                }
            }
            if (isICellElement(v)) {
                return <td className={v.className ? v.className : ""} key={i}>{v.value}</td>;
            } else {
                return <td key={i}>{v}</td>;
            }
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

    private handleRowClick(item: T) {
        if (this.props.onRowClick) {
            this.props.onRowClick(item);
        }
    }
}
