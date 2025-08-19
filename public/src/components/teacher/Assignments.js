import React, { useState } from "react";
import { isManuallyGraded, Color, hasBenchmarks, hasCriteria } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import Button, { ButtonType } from "../admin/Button";
import EditBenchmark from "./EditBenchmark";
import EditCriterion from "./EditCriterion";
import { useCourseID } from "../../hooks/useCourseID";
const Assignments = () => {
    const courseID = useCourseID();
    const actions = useActions().global;
    const state = useAppState();
    const AssignmentElement = ({ assignment }) => {
        const [hidden, setHidden] = useState(false);
        const [buttonText, setButtonText] = useState("Rebuild all tests");
        const rebuild = async () => {
            if (confirm(`Warning! This will rebuild all submissions for ${assignment.name}. This may take several minutes. Are you sure you want to continue?`)) {
                setButtonText("Rebuilding...");
                const success = await actions.rebuildAllSubmissions({ assignmentID: assignment.ID, courseID: courseID });
                if (success) {
                    setButtonText("Finished rebuilding");
                }
                else {
                    setButtonText("Failed to rebuild");
                }
            }
        };
        const assignmentForm = hasBenchmarks(assignment) ? assignment.gradingBenchmarks.map((bm) => (React.createElement(EditBenchmark, { key: bm.ID.toString(), benchmark: bm, assignment: assignment },
            hasCriteria(bm) && bm.criteria?.map((crit) => (React.createElement(EditCriterion, { key: crit.ID.toString(), originalCriterion: crit, assignment: assignment, benchmarkID: bm.ID }))),
            React.createElement(EditCriterion, { key: bm.criteria.length, assignment: assignment, benchmarkID: bm.ID })))) : null;
        return (React.createElement("ul", { key: assignment.ID.toString(), className: "list-group" },
            React.createElement("div", { onClick: () => setHidden(!hidden), role: "button", "aria-hidden": "true" },
                React.createElement("li", { key: "assignment", className: "list-group-item" }, assignment.name)),
            hidden && (React.createElement("li", { key: "form", className: "list-group-item" }, isManuallyGraded(assignment.reviewers)
                ? React.createElement(React.Fragment, null,
                    " ",
                    assignmentForm,
                    " ",
                    React.createElement(EditBenchmark, { key: assignment.gradingBenchmarks.length, assignment: assignment }))
                : React.createElement(Button, { text: buttonText, color: Color.BLUE, type: ButtonType.BUTTON, onClick: rebuild })))));
    };
    return (React.createElement("div", { className: "column" }, state.assignments[courseID.toString()]?.map(assignment => React.createElement(AssignmentElement, { key: assignment.ID, assignment: assignment }))));
};
export default Assignments;
