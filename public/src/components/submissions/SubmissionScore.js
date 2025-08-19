import React from "react";
const SubmissionScore = ({ score, totalWeight, }) => {
    const className = score.Score === score.MaxScore ? "passed" : "failed";
    const percentage = (score.Score / score.MaxScore) * (score.Weight / totalWeight) * 100;
    const maxPercentage = (score.MaxScore / score.MaxScore) * (score.Weight / totalWeight) * 100;
    const cellColor = percentage === maxPercentage ? "text-success" : "text-danger";
    return (React.createElement("tr", null,
        React.createElement("td", { className: `${className} pl-4` }, score.TestName),
        React.createElement("td", { className: "fixed-width-score" },
            score.Score,
            "/",
            score.MaxScore),
        React.createElement("td", { className: "fixed-width-percent" },
            React.createElement("span", { className: cellColor },
                percentage.toFixed(1),
                "%")),
        React.createElement("td", { className: "fixed-width-percent" },
            React.createElement("span", { style: { opacity: 0.5 }, title: `Weight: ${score.Weight}`, "aria-label": `Max weighted percentage is ${maxPercentage.toFixed(1)} percent, weight ${score.Weight}` },
                maxPercentage.toFixed(1),
                "%"))));
};
export default SubmissionScore;
