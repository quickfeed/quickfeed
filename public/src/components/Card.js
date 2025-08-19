import React from "react";
import { useNavigate } from "react-router";
const Card = (props) => {
    const navigate = useNavigate();
    const notification = props.notification
        ? React.createElement("i", { className: `badge badge-${props.notification.color} float-right` },
            " ",
            props.notification.text,
            " ")
        : null;
    const onClick = () => {
        if (props.onclick) {
            props.onclick();
        }
        else if (props.to) {
            navigate(props.to);
        }
    };
    return (React.createElement("div", { className: "col-sm-6", style: { marginBottom: "10px" } },
        React.createElement("div", { className: "card" },
            React.createElement("div", { className: "card-body" },
                React.createElement("h5", { className: "card-title" },
                    props.title,
                    " ",
                    notification),
                React.createElement("p", { className: "card-text" }, props.text),
                React.createElement("div", { className: "btn btn-primary", onClick: onClick }, props.buttonText)))));
};
export default Card;
