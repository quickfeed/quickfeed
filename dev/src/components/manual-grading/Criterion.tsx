import React, { useState } from "react"
import { GradingCriterion } from "../../../proto/ag/ag_pb"
import { useAppState } from "../../overmind"
import GradeComment from "./GradeComment"
import CriteriaStatus from "./CriteriaStatus"

/* Criteria component for the manual grading page */
const Criteria = ({ criteria }: { criteria: GradingCriterion }): JSX.Element => {

    // editing, setEditing is used to toggle the GradeComment component
    const [editing, setEditing] = useState<boolean>(false)
    const { isTeacher } = useAppState()

    // classname is used to style the first column of the row returned by this component
    // it adds a vertical line to the left of the row with color based on the grading criterion.
    let className: string
    switch (criteria.getGrade()) {
        case GradingCriterion.Grade.PASSED:
            className = "passed"
            break
        case GradingCriterion.Grade.FAILED:
            className = "failed"
            break
        case GradingCriterion.Grade.NONE:
            className = "not-graded"
            break
    }

    const passed = criteria.getGrade() == GradingCriterion.Grade.PASSED
    // manageOrShowPassed renders the ManageCriteriaStatus component if the user is a teacher, otherwise it renders a passed/failed icon
    const criteriaStatusOrPassFailIcon = isTeacher
        ? <CriteriaStatus criterion={criteria} />
        : <i className={passed ? "fa fa-check" : "fa fa-exclamation-circle"}></i>

    return (
        <>
            <tr className="align-items-center">
                <th className={className}>{criteria.getDescription()}</th>
                <th>
                    {criteriaStatusOrPassFailIcon}
                </th>
                <th onClick={() => setEditing(true)}>{criteria.getComment()}</th>
            </tr>
            <GradeComment grade={criteria} editing={editing} setEditing={setEditing} />
        </>
    )
}

export default Criteria
