import React from "react"
import { GradingBenchmark } from "../../../proto/qf/types_pb"
import { hasCriteria } from "../../Helpers"

/** RubricDisplay component shows the grading rubric for a given benchmark
 * with its criteria in a styled card format */
const RubricDisplay = ({ benchmark }: { benchmark: GradingBenchmark }) => {
    return (
        <div className="card mb-3 shadow-sm">
            <div className="card-header bg-primary text-white">
                <h5 className="mb-0 font-weight-bold">{benchmark.heading}</h5>
            </div>
            {hasCriteria(benchmark) && (
                <div className="card-body p-0">
                    <ul className="list-group list-group-flush">
                        {benchmark.criteria.map((criterion, index) => (
                            <li
                                key={criterion.ID.toString()}
                                className={`list-group-item ${index % 2 === 0 ? "bg-light" : ""
                                    }`}
                            >
                                <div className="d-flex justify-content-between align-items-center">
                                    <div className="flex-grow-1">
                                        <span className="text-dark">{criterion.description}</span>
                                    </div>
                                    {criterion.points > 0n && (
                                        <span className="badge badge-pill badge-primary ml-3 px-3 py-2">
                                            {criterion.points.toString()} pts
                                        </span>
                                    )}
                                </div>
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    )
}

export default RubricDisplay
