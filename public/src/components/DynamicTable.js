import React, { memo } from "react";
import { isHidden } from "../Helpers";
import { useAppState } from "../overmind";
const isCellElement = (element) => {
    return element.value !== undefined;
};
const isJSXElement = (element) => {
    return element.type !== undefined;
};
const DynamicTable = memo(({ header, data }) => {
    const [isMouseDown, setIsMouseDown] = React.useState(false);
    const container = React.useRef(null);
    const searchQuery = useAppState().query;
    if (!data || data.length === 0) {
        return null;
    }
    const isRowHidden = (row) => {
        if (searchQuery.length === 0) {
            return false;
        }
        for (const cell of row) {
            if (typeof cell === "string" && !isHidden(cell, searchQuery)) {
                return false;
            }
            if (isCellElement(cell) && !isHidden(cell.value, searchQuery)) {
                return false;
            }
            if (isJSXElement(cell)) {
                if (cell.props.hidden) {
                    return false;
                }
            }
        }
        return true;
    };
    const icon = (cell) => {
        return cell.iconClassName && cell.iconTitle ? React.createElement("i", { className: cell.iconClassName, title: cell.iconTitle }) : null;
    };
    const rowCell = (cell, index) => {
        if (isCellElement(cell)) {
            const element = cell.link ? React.createElement("a", { href: cell.link, target: "_blank", rel: "noopener noreferrer" }, cell.value) : cell.value;
            return React.createElement("td", { key: index, className: cell.className, onClick: cell.onClick },
                element,
                " ",
                icon(cell));
        }
        return index == 0 ? React.createElement("th", { key: index }, cell) : React.createElement("td", { key: index }, cell);
    };
    const headerRowCell = (cell, index) => {
        if (isCellElement(cell)) {
            const element = cell.link ? React.createElement("a", { href: cell.link }, cell.value) : cell.value;
            const style = cell.onClick ? { "cursor": "pointer" } : undefined;
            return React.createElement("th", { key: index, className: cell.className, style: style, onClick: cell.onClick },
                element,
                " ",
                icon(cell));
        }
        return React.createElement("th", { key: index }, cell);
    };
    const head = header.map((cell, index) => { return headerRowCell(cell, index); });
    const rows = data.map((row, index) => {
        const generatedRow = row.map((cell, index) => {
            return rowCell(cell, index);
        });
        return React.createElement("tr", { hidden: isRowHidden(row), key: index }, generatedRow);
    });
    const onMouseDown = () => {
        setIsMouseDown(true);
    };
    const onMouseMove = (e) => {
        e.preventDefault();
        if (!isMouseDown) {
            return;
        }
        if (container.current) {
            container.current.scrollLeft = container.current.scrollLeft - e.movementX;
        }
    };
    const onMouseUp = () => {
        setIsMouseDown(false);
    };
    return (React.createElement("div", { className: "table-overflow", ref: container, onMouseDown: onMouseDown, onMouseMove: onMouseMove, onMouseUp: onMouseUp, onMouseLeave: onMouseUp, role: "button", "aria-hidden": "true" },
        React.createElement("table", { className: "table table-striped table-grp" },
            React.createElement("thead", { className: "thead-dark" },
                React.createElement("tr", null, head)),
            React.createElement("tbody", null, rows))));
});
DynamicTable.displayName = "DynamicTable";
export default DynamicTable;
