import React from "react";
import { isVisible } from "../Helpers";
import { useActions, useAppState } from "../overmind";
const CourseFavoriteButton = ({ enrollment, style }) => {
    const actions = useActions().global;
    useAppState();
    const starIcon = isVisible(enrollment) ? 'fa fa-star' : "fa fa-star-o";
    return (React.createElement("span", { style: style, title: "Favorite or unfavorite this course. Favorite courses will appear on your dashboard." },
        React.createElement("i", { role: "button", "aria-hidden": "true", className: starIcon, onClick: () => actions.setEnrollmentState(enrollment) })));
};
export default CourseFavoriteButton;
