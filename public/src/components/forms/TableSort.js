import React, { useEffect } from "react";
import { SubmissionSort } from "../../Helpers";
import { useActions, useAppState } from "../../overmind";
const TableSort = ({ review }) => {
    const state = useAppState();
    const actions = useActions().global;
    useEffect(() => {
        return () => {
            actions.setSubmissionSort(SubmissionSort.Approved);
            actions.setAscending(true);
            actions.clearSubmissionFilter();
        };
    }, [actions]);
    const handleChange = (sort) => {
        actions.setSubmissionSort(sort);
    };
    const toggleIndividualSubmissions = () => {
        actions.setIndividualSubmissionsView(!state.individualSubmissionView);
    };
    const boldText = (sort) => {
        return state.sortSubmissionsBy === sort ? "font-weight-bold" : "";
    };
    const pointer = state.sortAscending ? "fa fa-caret-down" : "fa fa-caret-down fa-rotate-180";
    const textForToggleIndividualViewButton = state.individualSubmissionView ? "Individual" : "Group";
    const submissionFilters = [
        { name: "teachers", text: "Teachers", show: true },
        { name: "approved", text: "Graded", show: true },
        { name: "released", text: "Released", show: review }
    ];
    const filterElements = submissionFilters.map((filter) => {
        const displayText = state.submissionFilters.includes(filter.name)
            ? React.createElement("del", null, filter.text)
            : filter.text;
        return filter.show
            ? React.createElement(DivButton, { key: filter.name, text: displayText, onclick: () => actions.setSubmissionFilter(filter.name) })
            : null;
    });
    const sortByButtons = [
        { key: "approved", text: "Approved", className: boldText(SubmissionSort.Approved), onclick: () => handleChange(SubmissionSort.Approved) },
        { key: "score", text: "Score", className: boldText(SubmissionSort.Score), onclick: () => handleChange(SubmissionSort.Score) },
        { key: "pointer", text: React.createElement("i", { className: pointer }), onclick: () => actions.setAscending(!state.sortAscending) }
    ];
    const sortByElements = sortByButtons.map((button) => (React.createElement(DivButton, { key: button.key, text: button.text, className: button.className, onclick: button.onclick })));
    return (React.createElement("div", { className: "p-1 mb-2 bg-dark text-white d-flex flex-row" },
        React.createElement("div", { className: "d-inline-flex flex-row justify-content-center" },
            React.createElement("div", { className: "p-2" },
                React.createElement("span", null, "Sort by:")),
            sortByElements),
        React.createElement("div", { className: "d-inline-flex flex-row" },
            React.createElement("div", { className: "p-2" }, "Show:"),
            filterElements),
        React.createElement("div", { className: "d-inline-flex flex-row" },
            React.createElement(DivButton, { text: textForToggleIndividualViewButton, onclick: toggleIndividualSubmissions }))));
};
const DivButton = ({ text, key, className, onclick }) => {
    return (React.createElement("div", { key: key, className: `${className ?? ""} p-2`, role: "button", "aria-hidden": "true", onClick: onclick }, text));
};
export default TableSort;
