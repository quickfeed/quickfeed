import React from 'react';
const ScrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" });
};
const BackToTop = () => {
    return (React.createElement("footer", { className: "text-center mt-5" },
        React.createElement("button", { onClick: ScrollToTop, className: "btn align-items-center backToTop" },
            React.createElement("i", { className: "fa fa-arrow-up" }),
            React.createElement("p", null, "Back to top"))));
};
export default BackToTop;
