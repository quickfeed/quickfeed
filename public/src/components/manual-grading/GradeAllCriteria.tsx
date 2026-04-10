import React from "react"
import { useActions, useAppState } from "../../overmind"
import GradeButtons from "./GradeButtons"


const GradeAllCriteria = () => {
    const actions = useActions().review
    const { isTeacher, review } = useAppState()

    if (!isTeacher) {
        return null
    }

    return (
        <GradeButtons
            isActive={(grade) =>
                review.currentReview?.gradingBenchmarks.every(bm =>
                    bm.criteria.every(c => c.grade === grade)) ?? false
            }
            onClick={(grade) => actions.setAllGrade(grade)}
        />
    )
}

export default GradeAllCriteria
