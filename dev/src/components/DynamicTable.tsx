import React from "react"
import { isHidden } from "../Helpers"
import { useAppState } from "../overmind"

export type CellElement = {
    value: string,
    className?: string,
    onClick?: () => void,
    link?: string
}

const isCellElement = (obj: any): obj is CellElement => {
    return (obj as CellElement).value !== undefined
}


const DynamicTable = ({header, data}: {header: (string | JSX.Element)[], data: (string | JSX.Element | CellElement)[][]}) => {
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
        }
        return true
    }

    const head = header.map((value, index) => { return <th key={index}>{value}</th> })
    const rows = data.map((row, index) => {
        const generatedRow = row.map((cell, index) => {
            if (isCellElement(cell)) {
                const element = cell.link ? <a href={cell.link}>{cell.value}</a> : cell.value
                return <td key={index} className={cell.className} onClick={cell.onClick}>{element}</td>
            }
            return <td key={index}>{cell}</td>
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