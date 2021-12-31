import React, { useState } from "react"
import { GradingCriterion } from "../../../proto/ag/ag_pb"
import { useAppState } from "../../overmind"
import GradeComment from "./GradeComment"
import ManageCriteriaStatus from "./ManageCriteriaStatus"

const Criteria = ({criteria}: {criteria: GradingCriterion}): JSX.Element => {
    const [editing, setEditing] = useState<boolean>(false)
    const {isTeacher} = useAppState()
    let classname: string
    switch (criteria.getGrade()) {
        case GradingCriterion.Grade.PASSED:
            classname = "passed"
            break;
        case GradingCriterion.Grade.FAILED:
            classname = "failed"
            break;
        case GradingCriterion.Grade.NONE:
            classname  ="not-graded"

    }


    const passed = criteria.getGrade() == GradingCriterion.Grade.PASSED
    return (
        <>
        <tr className="align-items-center">
            <th className={classname}>{criteria.getDescription()}</th>
            <th> 
                { isTeacher ? <ManageCriteriaStatus criterion={criteria} /> : <i className={passed ? "fa fa-check" : "fa fa-exclamation-circle"}></i>}
            </th>
            <th onClick={() => setEditing(true)}>{criteria.getComment()}</th>
        </tr>
        <GradeComment grade={criteria} editing={editing} setEditing={setEditing} />
        </>
    )
}

export default Criteria