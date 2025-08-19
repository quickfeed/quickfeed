import React from "react";
import { hasBenchmarks } from "../Helpers";
import Benchmark from "./manual-grading/Benchmark";
import Criteria from "./manual-grading/Criterion";
import MarkReadyButton from "./manual-grading/MarkReadyButton";
import SummaryFeedback from "./manual-grading/SummaryFeedback";
const ReviewResult = ({ review }) => {
    const result = hasBenchmarks(review) ? review.gradingBenchmarks.map(benchmark => {
        return (React.createElement(Benchmark, { key: benchmark.ID.toString(), bm: benchmark }, benchmark.criteria.map(criteria => React.createElement(Criteria, { key: criteria.ID.toString(), criteria: criteria }))));
    }) : null;
    return (React.createElement("table", { className: "table" },
        React.createElement("thead", { className: "thead-dark" },
            React.createElement("tr", { className: "table-primary" },
                React.createElement("th", null, "Score:"),
                React.createElement("th", null, review.score),
                React.createElement("th", null)),
            React.createElement("tr", null,
                React.createElement("th", { scope: "col" }, "Criteria"),
                React.createElement("th", { scope: "col" }, "Status"),
                React.createElement("th", { scope: "col" }, "Comment"))),
        React.createElement("tbody", null, result),
        React.createElement("tfoot", null,
            React.createElement(SummaryFeedback, { review: review }),
            !review.ready
                ?
                    React.createElement("tr", null,
                        React.createElement(MarkReadyButton, { review: review }))
                : null)));
};
export default ReviewResult;
