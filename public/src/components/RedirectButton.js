import React from "react";
import { useLocation, useNavigate } from "react-router";
const RedirectButton = ({ to }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const isHidden = location.pathname === to;
    return (React.createElement("button", { className: "btn btn-dark redirectButton", type: "button", onClick: () => navigate(to), hidden: isHidden },
        React.createElement("i", { className: "fa fa-arrow-left" })));
};
export default RedirectButton;
