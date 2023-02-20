import React, { useState } from "react"
import { GradingCriterion, GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { useAppState } from "../../overmind"
import GradeComment from "./GradeComment"
import CriteriaStatus from "./CriteriaStatus"
import CriterionComment from "./Comment"
import UnstyledButton from "../UnstyledButton"


/* Criteria component for the manual grading page */
const Criteria = ({ criteria }: { criteria: GradingCriterion }): JSX.Element => {

    // editing, setEditing is used to toggle the GradeComment component
    const [editing, setEditing] = useState<boolean>(false)
    const [showComment, setShowComment] = React.useState<boolean>(true)
    const { isTeacher } = useAppState()

    // classname is used to style the first column of the row returned by this component
    // it adds a vertical line to the left of the row with color based on the grading criterion.
    let className: string
    switch (criteria.grade) {
        case GradingCriterion_Grade.PASSED:
            className = "passed"
            break
        case GradingCriterion_Grade.FAILED:
            className = "failed"
            break
        case GradingCriterion_Grade.NONE:
            className = "not-graded"
            break
    }

    const passed = criteria.grade == GradingCriterion_Grade.PASSED
    // manageOrShowPassed renders the ManageCriteriaStatus component if the user is a teacher, otherwise it renders a passed/failed icon
    const criteriaStatusOrPassFailIcon = isTeacher
        ? <CriteriaStatus criterion={criteria} />
        : <i className={passed ? "fa fa-check" : "fa fa-exclamation-circle"} />


    let comment: JSX.Element
    if (isTeacher) {
        // Display edit icon if comment is empty
        // If comment is not empty, display the comment
        if (criteria.comment.length > 0) {
            comment = <CriterionComment comment={criteria.comment} />
        } else {
            comment = <i style={{ opacity: "0.5" }} className="fa fa-pencil-square-o" aria-hidden="true" />
        }
    } else {
        comment = <CriterionComment comment={criteria.comment} />
    }

    // Only display the comment if the user is a teacher or if the comment is not empty
    const displayComment = isTeacher ||  criteria.comment.length > 0
    return (
        <>
            <tr className="align-items-center">
                <td className={className}>{criteria.description}</td>
                <td>
                    {criteriaStatusOrPassFailIcon}
                </td>
                <td>
                    {displayComment ? <UnstyledButton onClick={() => setShowComment(!showComment)}><i className={`fa fa-comment${!showComment ? "-o" : ""}`} /></UnstyledButton> : null}
                </td>
            </tr>
            {displayComment ?
            <tr className={`comment comment-${className}${!showComment ? " hidden" : "" } `}>
                <td onClick={() => setEditing(true)} colSpan={3}>
                    {comment}
                </td>
            </tr> : null
            }
            <GradeComment grade={criteria} editing={editing} setEditing={setEditing} />
        </>
    )
}

export default Criteria
