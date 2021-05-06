import React from "react"
import { useOvermind } from "../overmind"
import { GradingCriterion, Review } from "../proto/ag_pb"


interface submission {
    review: Review[]
}

const ReviewResult = ({review}: submission) => {
    const {state: {theme}} = useOvermind()
    const result = (): JSX.Element[] => {
        let b: JSX.Element[] = []

        review.forEach(r => {
            r.getBenchmarksList().map(benchmark => {
                b.push(
                <tr className="table-info">
                    <th colSpan={2}>{benchmark.getHeading()}</th>
                    <th>{benchmark.getComment()}</th>
                </tr>)
                benchmark.getCriteriaList().map(criteria => {
                    b.push(
                        <tr>
                            <th style={{boxShadow: criteria.getGrade() == GradingCriterion.Grade.PASSED ? "0 0px 0 #000 inset, 5px 0 0 green inset" :"0 0px 0 #000 inset, 5px 0 0 red inset"}}>{criteria.getDescription()} {criteria.getComment()}</th>
                            <th><i className={criteria.getGrade() == GradingCriterion.Grade.PASSED ? "fa fa-check" : "fa fa-exclamation-circle"}></i></th>
                            <th>{criteria.getComment()}</th>
                        </tr>)
                })
            })
            
        })
        return b
    }

    return (
        <div className="container">
            <table className="table"> 
                <thead className={theme == "light" ? "thead-dark" : "thead-light"}>
                    <tr className="table-primary">
                        <th>{review[0].getFeedback()}</th>
                        <th>{review[0].getScore()}%</th>
                        <th></th>
                    </tr>
                    <tr>
                        <th scope="col">Criteria</th>
                        <th scope="col">Status</th>
                        <th scope="col">Comment</th>
                    </tr>
                </thead>
                    {result()}
            </table>    

        </div>
    )
}

export default ReviewResult