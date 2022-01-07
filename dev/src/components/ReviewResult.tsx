import { json } from "overmind"
import React from "react"
import { Review } from "../../proto/ag/ag_pb"
import Benchmark from "./manual-grading/Benchmark"
import Criteria from "./manual-grading/Criterion"

const ReviewResult = ({ review }: { review?: Review }): JSX.Element => {

    const result = json(review)?.getGradingbenchmarksList().map((benchmark, index) => {
        return (
            <Benchmark key={index} bm={benchmark}>
                {benchmark.getCriteriaList().map((criteria, index) => <Criteria key={index} criteria={criteria} />)}
            </Benchmark>
        )
    })

    return (
        <div>
            {review &&
                <table className="table">
                    <thead className="thead-dark">
                        {review &&
                            <tr className="table-primary">
                                <th>{review.getFeedback()}</th>
                                <th>{review.getScore()}</th>
                                <th></th>
                            </tr>
                        }
                        <tr>
                            <th scope="col">Criteria</th>
                            <th scope="col">Status</th>
                            <th scope="col">Comment</th>
                        </tr>
                    </thead>
                    <tbody>
                        {result}
                    </tbody>
                </table>
            }

        </div>
    )
}

export default ReviewResult