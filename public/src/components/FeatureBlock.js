import React from "react";
const FeatureBlock = ({ heading, subheading, content, imageSrc, reverse = false, }) => {
    return (React.createElement("div", { className: `row featurette ${reverse ? "flex-row-reverse" : ""}` },
        React.createElement("div", { className: "col-md-7" },
            React.createElement("h2", { className: "featurette-heading" },
                heading,
                ": ",
                React.createElement("span", { className: "text-muted" }, subheading)),
            React.createElement("p", { className: "lead" }, content)),
        React.createElement("div", { className: "col-md-5" },
            React.createElement("img", { className: "featurette-image img-responsive about", src: imageSrc, alt: "Feature example" }))));
};
export const MiniFeatureBlock = ({ title, content, media, style, }) => {
    return (React.createElement("div", { className: "col-lg-4 text-center" },
        React.createElement("div", { style: {
                width: "140px",
                height: "140px",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                margin: "0 auto",
                ...style
            } }, media),
        React.createElement("h2", null, title),
        React.createElement("p", null, content)));
};
export default FeatureBlock;
