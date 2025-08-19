import React, { useState } from "react";
import GradeComment from "./GradeComment";
const Benchmark = ({ children, bm }) => {
    const [editing, setEditing] = useState(false);
    return (React.createElement(React.Fragment, null,
        React.createElement("tr", { className: "table-info" },
            React.createElement("th", { colSpan: 2 }, bm.heading),
            React.createElement("th", { onClick: () => setEditing(true) }, bm.comment)),
        React.createElement(GradeComment, { grade: bm, editing: editing, setEditing: setEditing }),
        children));
};
export default Benchmark;
