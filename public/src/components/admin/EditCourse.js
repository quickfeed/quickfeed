import React, { useState } from "react";
import { useAppState } from "../../overmind";
import DynamicTable from "../DynamicTable";
import CourseForm from "../forms/CourseForm";
const EditCourse = () => {
    const state = useAppState();
    const [course, setCourse] = useState();
    const courses = state.courses.map(c => {
        const selected = course?.ID === c.ID;
        const badge = selected ? "badge badge-danger" : "badge badge-primary";
        const buttonText = selected ? "Cancel" : "Edit";
        return [
            c.name,
            c.code,
            c.tag,
            c.year.toString(),
            c.slipDays.toString(),
            React.createElement("button", { key: c.ID, className: `clickable ${badge}`, onClick: () => setCourse(selected ? undefined : c) }, buttonText)
        ];
    });
    return (React.createElement("div", { className: "box" },
        React.createElement(DynamicTable, { header: ["Course", "Code", "Tag", "Year", "Slipdays", "Edit"], data: courses }),
        course ? React.createElement(CourseForm, { courseToEdit: course }) : null));
};
export default EditCourse;
