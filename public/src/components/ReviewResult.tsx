import React from "react"
import { Review } from "../../gen/qf/types_pb"
import { hasBenchmarks } from "../Helpers"
import Benchmark from "./manual-grading/Benchmark"
import Criteria from "./manual-grading/Criterion"
import SummaryFeedback from "./manual-grading/SummaryFeedback"


const ReviewResult = ({ review }: { review?: Review }): JSX.Element | null => {

    if (!review) {
        return null
    }

    const result = hasBenchmarks(review) ? review.gradingBenchmarks.map((benchmark, index) => {
        return (
            <Benchmark key={index} bm={benchmark}>
                {benchmark.criteria.map((criteria, index) => <Criteria key={index} criteria={criteria} />)}
            </Benchmark>
        )
    }) : null

    return (
        <table className="table">
            <thead className="thead-dark">
                <tr className="table-primary">
                    <th>Score:</th>
                    <th>{review.score}</th>
                    <th />
                </tr>
                <tr>
                    <th scope="col">Criteria</th>
                    <th scope="col">Status</th>
                    <th scope="col">Comment</th>
                </tr>
            </thead>
            <tbody>
                {result}
            </tbody>
            <tfoot>
                <SummaryFeedback review={review} />
            </tfoot>
        </table>
    )
}

export default ReviewResult
