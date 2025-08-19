import React, { useCallback, useState } from "react";
import { hasEnrollment } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
import FormInput from "../forms/FormInput";
import { useNavigate } from "react-router";
import { clone } from "@bufbuild/protobuf";
import { UserSchema } from "../../../proto/qf/types_pb";
const ProfileForm = ({ children, setEditing }) => {
    const state = useAppState();
    const actions = useActions().global;
    const navigate = useNavigate();
    const [user, setUser] = useState(clone(UserSchema, state.self));
    const [isValid, setIsValid] = useState(state.isValid);
    const handleChange = useCallback((event) => {
        const { name, value } = event.currentTarget;
        switch (name) {
            case "name":
                user.Name = value;
                break;
            case "email":
                user.Email = value;
                break;
            case "studentid":
                user.StudentID = value;
                break;
        }
        setUser(user);
        if (user.Name !== "" && user.Email !== "" && user.StudentID !== "") {
            setIsValid(true);
        }
        else {
            setIsValid(false);
        }
    }, [user]);
    const submitHandler = (e) => {
        e.preventDefault();
        actions.updateUser(user);
        setEditing(false);
        if (!hasEnrollment(state.enrollments)) {
            navigate("/courses");
        }
    };
    return (React.createElement("div", null,
        !isValid ? children : null,
        React.createElement("form", { className: "form-group", onSubmit: submitHandler },
            React.createElement(FormInput, { prepend: "Name", name: "name", defaultValue: user.Name, onChange: handleChange }),
            React.createElement(FormInput, { prepend: "Email", name: "email", defaultValue: user.Email, onChange: handleChange, type: "email" }),
            React.createElement(FormInput, { prepend: "Student ID", name: "studentid", defaultValue: user.StudentID, onChange: handleChange, type: "number" }),
            React.createElement("div", { className: "col input-group mb-3" },
                React.createElement("input", { className: "btn btn-primary", disabled: !isValid, type: "submit", value: "Save", style: { marginTop: "20px" } })))));
};
export default ProfileForm;
