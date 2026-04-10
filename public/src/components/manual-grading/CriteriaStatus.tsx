import React from "react"
import { GradingCriterion } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import GradeButtons from "./GradeButtons"


const CriteriaStatus = ({ criterion }: { criterion: GradingCriterion }) => {
    const { setGrade } = useActions().review
    const { isTeacher } = useAppState()

    if (!isTeacher) {
        return null
    }

    return (
        <GradeButtons
            isActive={(grade) => criterion.grade === grade}
            onClick={(grade) => setGrade({ criterion, grade })}
        />
    )
}

export default CriteriaStatus
