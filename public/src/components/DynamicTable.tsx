import React from "react"
import { isHidden } from "../Helpers"
import { useAppState } from "../overmind"

export type CellElement = {
    value: string,
    className?: string,
    onClick?: () => void,
    link?: string
}

export type RowElement = (string | JSX.Element | CellElement)
export type Row = RowElement[]

const isCellElement = (element: RowElement): element is CellElement => {
    return (element as CellElement).value !== undefined
}

const isJSXElement = (element: RowElement): element is JSX.Element => {
    return (element as JSX.Element).type !== undefined
}

const DynamicTable = ({ header, data }: { header: Row, data: Row[] }): JSX.Element | null => {
    if (!data || data.length === 0) {
        // Nothing to render
        return null
    }

    const searchQuery = useAppState().query

    const isRowHidden = (row: Row) => {
        if (searchQuery.length === 0) {
            return false
        }
        for (const cell of row) {
            if (typeof cell === "string" && !isHidden(cell, searchQuery)) {
                return false
            }
            if (isCellElement(cell) && !isHidden(cell.value, searchQuery)) {
                return false
            }
            // To enable searching with JSX.Element, add a 'hidden: boolean' prop to the element
            if (isJSXElement(cell)) {
                if (cell.props.hidden) {
                    return false
                }
            }
        }
        return true
    }

    const rowCell = (cell: RowElement, index: number) => {
        if (isCellElement(cell)) {
            const element = cell.link ? <a href={cell.link} target={"_blank"} rel="noopener noreferrer">{cell.value}</a> : cell.value
            return <td key={index} className={cell.className} onClick={cell.onClick}>{element}</td>
        }
        return index == 0 ? <th key={index}>{cell}</th> : <td key={index}>{cell}</td>
    }

    const headerRowCell = (cell: RowElement, index: number) => {
        if (isCellElement(cell)) {
            const element = cell.link ? <a href={cell.link}>{cell.value}</a> : cell.value
            return <th key={index} className={cell.className} style={cell.onClick ? { "cursor": "pointer" } : undefined} onClick={cell.onClick}>{element}</th>
        }
        return <th key={index}>{cell}</th>
    }

    const head = header.map((cell, index) => { return headerRowCell(cell, index) })

    const rows = data.map((row, index) => {
        const generatedRow = row.map((cell, index) => {
            return rowCell(cell, index)
        })
        return <tr hidden={isRowHidden(row)} key={index}>{generatedRow}</tr>
    })

    return (
        <table className="table table-striped table-grp">
            <thead className="thead-dark">
                <tr>
                    {head}
                </tr>
            </thead>
            <tbody>
                {rows}
            </tbody>
        </table>
    )
}

export default DynamicTable
