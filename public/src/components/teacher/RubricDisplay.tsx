import React from "react"
import { GradingBenchmark } from "../../../proto/qf/types_pb"
import { hasCriteria } from "../../Helpers"

/** RubricDisplay component shows the grading rubric for a given benchmark
 * with its criteria in a styled card format */
const RubricDisplay = ({ benchmark }: { benchmark: GradingBenchmark }) => {
    return (
        <div className="card bg-base-100 shadow-sm mb-3">
            <div className="card-title bg-primary text-primary-content px-4 py-3 rounded-t-2xl">
                <h5 className="font-bold">{benchmark.heading}</h5>
            </div>
            {hasCriteria(benchmark) && (
                <ul className="divide-y divide-base-content/10">
                    {benchmark.criteria.map((criterion, index) => (
                        <li
                            key={criterion.ID.toString()}
                            className={`flex items-center justify-between px-4 py-3 ${index % 2 === 0 ? "bg-base-200" : "bg-base-100"}`}
                        >
                            <span>{criterion.description}</span>
                            {criterion.points > 0n && (
                                <span className="badge badge-primary ml-3 px-3 py-2 shrink-0">
                                    {criterion.points.toString()} pts
                                </span>
                            )}
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}

export default RubricDisplay
