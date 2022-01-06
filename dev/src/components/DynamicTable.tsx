import React, { useEffect } from "react"
import { isHidden } from "../Helpers"
import { useAppState } from "../overmind"

export type CellElement = {
    value: string,
    className?: string,
    onClick?: () => void,
    link?: string
}

const isCellElement = (obj: unknown): obj is CellElement => {
    return (obj as CellElement).value !== undefined
}

const isJSXElement = (obj: unknown): obj is JSX.Element => {
    return (obj as JSX.Element).type !== undefined
}

const DynamicTable = ({ header, data }: { header: (string | JSX.Element | CellElement)[], data: (string | JSX.Element | CellElement)[][] }): JSX.Element | null => {
    // If there is no data, don't render anything
    if (!data || data.length === 0) {
        return null
    }

    const searchQuery = useAppState().query

    const isRowHidden = (row: (string | JSX.Element | CellElement)[]) => {
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

    const generateRow = (cell: string | JSX.Element | CellElement | (string | JSX.Element | CellElement)[], index: number) => {
        if (isCellElement(cell)) {
            const element = cell.link ? <a href={cell.link}>{cell.value}</a> : cell.value
            return <td key={index} className={cell.className} onClick={cell.onClick}>{element}</td>
        }
        return index == 0 ? <th key={index}>{cell}</th> : <td key={index}>{cell}</td>
    }

    const generateHeaderRow = (cell: string | JSX.Element | CellElement, index: number) => {
        if (isCellElement(cell)) {
            const element = cell.link ? <a href={cell.link}>{cell.value}</a> : cell.value
            return <th key={index} className={cell.className} style={cell.onClick ? { "cursor": "pointer" } : undefined} onClick={cell.onClick}>{element}</th>
        }
        return <th key={index}>{cell}</th>
    }

    const head = header.map((cell, index) => { return generateHeaderRow(cell, index) })

    const rows = data.map((row, index) => {
        const generatedRow = row.map((cell, index) => {
            return generateRow(cell, index)
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