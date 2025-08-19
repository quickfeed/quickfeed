import React, { useState } from "react";
import { GradingBenchmarkSchema } from "../../../proto/qf/types_pb";
import { useActions } from "../../overmind";
import { clone, create } from "@bufbuild/protobuf";
const EditBenchmark = ({ children, benchmark, assignment }) => {
    const actions = useActions().global;
    const [editing, setEditing] = useState(false);
    const [add, setAdd] = useState(benchmark ? false : true);
    const bm = benchmark
        ? clone(GradingBenchmarkSchema, benchmark)
        : create(GradingBenchmarkSchema);
    const resetBenchmark = () => {
        bm.heading = "";
        setAdd(true);
    };
    const handleBenchmark = async (event) => {
        const { value } = event.currentTarget;
        if (event.key === "Enter") {
            bm.AssignmentID = assignment.ID;
            const success = await actions.createOrUpdateBenchmark({ benchmark: bm, assignment });
            if (!success) {
                resetBenchmark();
            }
            setEditing(false);
        }
        else {
            bm.heading = value;
        }
    };
    const handleBlur = () => {
        if (benchmark) {
            bm.heading = benchmark.heading;
        }
        else {
            resetBenchmark();
        }
        setEditing(false);
    };
    if (add) {
        return (React.createElement("div", { className: "list-group-item list-group-item-primary" },
            React.createElement("button", { className: "btn btn-primary", name: "submit", onClick: () => { setAdd(false); setEditing(true); } }, "Add Benchmark")));
    }
    const input = React.createElement("input", { className: "form-control", type: "text", autoFocus: true, defaultValue: bm?.heading, onBlur: handleBlur, onKeyUp: e => handleBenchmark(e) });
    const textAndButton = (React.createElement("span", { onClick: () => setEditing(!editing), role: "button", "aria-hidden": "true" },
        bm?.heading,
        React.createElement("button", { className: "p-2 badge badge-danger float-right clickable", onClick: () => actions.deleteBenchmark({ benchmark, assignment }) }, "Delete Benchmark")));
    return (React.createElement(React.Fragment, null,
        React.createElement("div", { className: "list-group-item list-group-item-primary" }, editing ? input : textAndButton),
        children));
};
export default EditBenchmark;
