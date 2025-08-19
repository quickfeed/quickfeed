import React from "react";
import ReactMarkdown from "react-markdown";
const CriterionComment = ({ comment }) => {
    if (comment == "" || comment.length == 0) {
        return null;
    }
    return (React.createElement("div", { className: "comment-md" },
        React.createElement(ReactMarkdown, { children: comment, components: {
                code({ node, className, children, ref, ...props }) {
                    return (React.createElement("code", { className: className, ...props }, children));
                }
            } })));
};
export default CriterionComment;
