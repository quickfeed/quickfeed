import React from "react";
const Loading = () => {
    return (React.createElement("div", { className: "centered" },
        React.createElement("i", { className: "fa fa-refresh fa-spin fa-3x fa-fw" }),
        React.createElement("p", null,
            React.createElement("strong", null, "Loading..."))));
};
export default Loading;
