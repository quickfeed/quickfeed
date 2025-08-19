import React, { useCallback, useState } from "react";
import { useActions } from "../../overmind";
import { CourseSchema } from "../../../proto/qf/types_pb";
import FormInput from "./FormInput";
import { useNavigate } from "react-router";
import { clone } from "@bufbuild/protobuf";
const CourseForm = ({ courseToEdit }) => {
    const actions = useActions().global;
    const navigate = useNavigate();
    const [course, setCourse] = useState(clone(CourseSchema, courseToEdit));
    const handleChange = useCallback((event) => {
        const { name, value } = event.currentTarget;
        switch (name) {
            case "courseName":
                course.name = value;
                break;
            case "courseTag":
                course.tag = value;
                break;
            case "courseCode":
                course.code = value;
                break;
            case "courseYear":
                course.year = Number(value);
                break;
            case "slipDays":
                course.slipDays = Number(value);
                break;
        }
        setCourse(course);
    }, [course]);
    const submitHandler = async (e) => {
        e.preventDefault();
        await actions.editCourse({ course });
        navigate(`/course/${course.ID}`);
    };
    return (React.createElement("div", { className: "container" },
        React.createElement("form", { className: "form-group", onSubmit: async (e) => await submitHandler(e) },
            React.createElement("div", { className: "row" },
                React.createElement(FormInput, { prepend: "Name", name: "courseName", placeholder: "Course Name", defaultValue: course.name, onChange: handleChange })),
            React.createElement("div", { className: "row" },
                React.createElement(FormInput, { prepend: "Code", name: "courseCode", placeholder: "(ex. DAT320)", defaultValue: course.code, onChange: handleChange }),
                React.createElement(FormInput, { prepend: "Tag", name: "courseTag", placeholder: "(ex. Fall / Spring)", defaultValue: course.tag, onChange: handleChange })),
            React.createElement("div", { className: "row" },
                React.createElement(FormInput, { prepend: "Slip days", name: "slipDays", placeholder: "(ex. 7)", defaultValue: course.slipDays.toString(), onChange: handleChange, type: "number" }),
                React.createElement(FormInput, { prepend: "Year", name: "courseYear", placeholder: "(ex. 2021)", defaultValue: course.year.toString(), onChange: handleChange, type: "number" })),
            React.createElement("input", { className: "btn btn-primary", type: "submit", value: "Save" }))));
};
export default CourseForm;
