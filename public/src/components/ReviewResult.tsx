import React from "react"
import { Review } from "../../proto/qf/types_pb"
import { hasBenchmarks } from "../Helpers"
import Benchmark from "./manual-grading/Benchmark"
import Criteria from "./manual-grading/Criterion"
import SummaryFeedback from "./manual-grading/SummaryFeedback"
import { useAppState } from "../overmind"
import GradeAllCriteria from "./manual-grading/GradeAllCriteria"


const ReviewResult = ({ review }: { review: Review }) => {
    const state = useAppState()
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
                {state.isTeacher ? <tr>
                    <td>Set all criteria to:</td>
                    <td><GradeAllCriteria /></td>
                    <td />
                </tr> : null}
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
