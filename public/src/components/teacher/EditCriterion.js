import React, { useState } from "react";
import { GradingCriterionSchema } from "../../../proto/qf/types_pb";
import { useActions } from "../../overmind";
import { clone, create } from "@bufbuild/protobuf";
const EditCriterion = ({ originalCriterion, benchmarkID, assignment }) => {
    const actions = useActions().global;
    const [editing, setEditing] = useState(false);
    const [add, setAdd] = useState(originalCriterion ? false : true);
    const criterion = originalCriterion
        ? clone(GradingCriterionSchema, originalCriterion)
        : create(GradingCriterionSchema);
    const resetCriterion = () => {
        criterion.description = "";
        setAdd(true);
    };
    const handleCriteria = async (event) => {
        const { value } = event.currentTarget;
        if (event.key === "Enter") {
            criterion.BenchmarkID = benchmarkID;
            const success = await actions.createOrUpdateCriterion({ criterion, assignment });
            if (!success) {
                resetCriterion();
            }
            setEditing(false);
        }
        else {
            criterion.description = value;
        }
    };
    const handleBlur = () => {
        if (originalCriterion) {
            criterion.description = originalCriterion.description;
        }
        else {
            resetCriterion();
        }
        setEditing(false);
    };
    if (add) {
        return (React.createElement("div", { className: "list-group-item" },
            React.createElement("button", { className: "btn btn-success", name: "submit", onClick: () => { setAdd(false); setEditing(true); } }, "Add Criteria")));
    }
    const input = React.createElement("input", { className: "form-control", type: "text", autoFocus: true, onBlur: handleBlur, defaultValue: criterion.description, name: "criterion", onKeyUp: e => handleCriteria(e) });
    const textAndButton = (React.createElement("span", { onClick: () => setEditing(!editing), role: "button", "aria-hidden": "true" },
        criterion.description,
        React.createElement("button", { className: "p-2 badge badge-danger float-right clickable", onClick: () => actions.deleteCriterion({ criterion: originalCriterion, assignment }) }, "Delete Criteria")));
    return (React.createElement("div", { className: "list-group-item" }, editing ? input : textAndButton));
};
export default EditCriterion;
