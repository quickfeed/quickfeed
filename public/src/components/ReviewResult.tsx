import React from "react"
import { Review } from "../../proto/qf/types_pb"
import { hasBenchmarks } from "../Helpers"
import Benchmark from "./manual-grading/Benchmark"
import Criteria from "./manual-grading/Criterion"
import MarkReadyButton from "./manual-grading/MarkReadyButton"
import SummaryFeedback from "./manual-grading/SummaryFeedback"


const ReviewResult = ({ review }: { review?: Review }) => {

    if (!review) {
        return null
    }

    const result = hasBenchmarks(review) ? review.gradingBenchmarks.map(benchmark => {
        return (
            <Benchmark key={benchmark.ID.toString()} bm={benchmark}>
                {benchmark.criteria.map(criteria => <Criteria key={criteria.ID.toString()} criteria={criteria} />)}
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
                {!review.ready
                    ?
                    <tr>
                        <MarkReadyButton review={review} />
                    </tr>
                    : null
                }
            </tfoot>
        </table>
    )
}

export default ReviewResult
