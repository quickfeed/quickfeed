import React from "react";
import { useAppState } from "../../overmind";
import AssignmentCard from "./AssignmentCard";
import { useCourseID } from "../../hooks/useCourseID";
const CourseLabs = () => {
    const state = useAppState();
    const courseID = useCourseID().toString();
    const assignments = state.assignments[courseID] || [];
    const selfID = state.self.ID;
    return (React.createElement("ul", { className: "list-group" }, assignments.map((assignment) => {
        const submissions = state.submissions.ForAssignment(assignment);
        return (React.createElement(AssignmentCard, { key: assignment.ID.toString(), assignment: assignment, submissions: submissions, courseID: courseID, selfID: selfID }));
    })));
};
export default CourseLabs;
