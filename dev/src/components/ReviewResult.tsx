import { json } from "overmind"
import React from "react"
import { GradingBenchmark, GradingCriterion, Review } from "../../proto/ag/ag_pb"
import ManageCriteriaStatus from "./ManageCriteriaStatus"

const ReviewResult = ({review, teacher}: {review?: Review, teacher?: boolean}): JSX.Element => {

    const Benchmark = ({children, bm}: {children: React.ReactNode, bm: GradingBenchmark}) => {
        return (
            <>
                <tr className="table-info">
                    <th colSpan={2}>{bm.getHeading()}</th>
                    <th>{bm.getComment()}</th>
                </tr>
                {children}
            </>
        )
    }

    const Criteria = ({criteria}: {criteria: GradingCriterion}) => {
        const passed = criteria.getGrade() == GradingCriterion.Grade.PASSED
        const boxShadow = passed ? "0 0px 0 #000 inset, 5px 0 0 green inset" : "0 0px 0 #000 inset, 5px 0 0 red inset"
        return (
            <tr className="align-items-center">
                <th style={{boxShadow: boxShadow}}>{criteria.getDescription()}</th>
                <th> { teacher ?
                    <ManageCriteriaStatus criterion={criteria} /> : <i className={passed ? "fa fa-check" : "fa fa-exclamation-circle"}></i>
                }
                </th>
                <th>{criteria.getComment()}</th>
            </tr>
        )
    }

    const result = json(review)?.getGradingbenchmarksList().map((benchmark, index) => {
        return (
            <Benchmark key={index} bm={benchmark}>
                {benchmark.getCriteriaList().map((criteria, index) => <Criteria key={index} criteria={criteria} />)}
            </Benchmark>
        )
    })


    // TODO: DynamicTable ?
    return (
        <div>
            { review &&
            <table className="table"> 
                <thead className="thead-dark">
                    {review &&
                    <tr className="table-primary">
                        <th>{review.getFeedback()}</th>
                        <th>{review.getScore()}%</th>
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