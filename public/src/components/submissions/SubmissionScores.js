import React, { useCallback } from 'react';
import SubmissionScore from "./SubmissionScore";
import { ScoreSchema } from "../../../proto/kit/score/score_pb";
import { clone } from "@bufbuild/protobuf";
const SubmissionScores = ({ submission }) => {
    const [sortKey, setSortKey] = React.useState("name");
    const [sortAscending, setSortAscending] = React.useState(true);
    const sortScores = () => {
        const sortBy = sortAscending ? 1 : -1;
        const scores = submission.Scores.map(score => clone(ScoreSchema, score));
        const totalWeight = scores.reduce((acc, score) => acc + score.Weight, 0);
        return scores.sort((a, b) => {
            switch (sortKey) {
                case "name":
                    return sortBy * (a.TestName.localeCompare(b.TestName));
                case "score":
                    return sortBy * (a.Score - b.Score);
                case "weight":
                    return sortBy * (a.Weight - b.Weight);
                case "percentage":
                    return sortBy * ((a.Score / a.MaxScore) * (a.Weight / totalWeight) - (b.Score / b.MaxScore) * (b.Weight / totalWeight));
                default:
                    return 0;
            }
        });
    };
    const handleSort = useCallback((event) => {
        const key = event.currentTarget.dataset.key;
        if (sortKey === key) {
            setSortAscending(!sortAscending);
        }
        else {
            setSortKey(key);
            setSortAscending(true);
        }
    }, [sortKey, sortAscending]);
    const sortedScores = React.useMemo(sortScores, [submission, sortKey, sortAscending]);
    const totalWeight = sortedScores.reduce((acc, score) => acc + score.Weight, 0);
    return (React.createElement("table", { className: "table table-curved table-striped table-hover" },
        React.createElement("thead", { className: "thead-dark" },
            React.createElement("tr", null,
                React.createElement("th", { colSpan: 1, className: "col-md-8", "data-key": "name", role: "button", onClick: handleSort }, "Test Name"),
                React.createElement("th", { colSpan: 1, className: "fixed-width-percent", "data-key": "score", role: "button", onClick: handleSort }, "Score"),
                React.createElement("th", { colSpan: 1, className: "fixed-width-percent", "data-key": "percentage", role: "button", onClick: handleSort }, "%"),
                React.createElement("th", { colSpan: 1, className: "fixed-width-percent", "data-key": "weight", "data-toggle": "tooltip", title: "Maximum % contribution to total score", role: "button", onClick: handleSort }, "Max"))),
        React.createElement("tbody", null, sortedScores.map(score => React.createElement(SubmissionScore, { key: score.ID.toString(), score: score, totalWeight: totalWeight }))),
        React.createElement("tfoot", null,
            React.createElement("tr", null,
                React.createElement("th", { colSpan: 2 }, "Total Score"),
                React.createElement("th", { className: "text-right" },
                    submission.score,
                    "%"),
                React.createElement("th", { className: "text-right" }, "100%")))));
};
export default SubmissionScores;
