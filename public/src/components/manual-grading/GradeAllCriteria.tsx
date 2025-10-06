import React from "react"
import { GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"


const GradeAllCriteria = () => {
    const actions = useActions().review
    const { isTeacher, review } = useAppState()

    if (!isTeacher) {
        return null
    }

    const handleSetAllGrade = (grade: GradingCriterion_Grade) => () => {
        actions.setAllGrade(grade)
    }

    const buttons: { icon: string, status: GradingCriterion_Grade, style: string, onClick: () => void }[] = [
        { icon: "fa fa-check", status: GradingCriterion_Grade.PASSED, style: "success", onClick: handleSetAllGrade(GradingCriterion_Grade.PASSED) },
        { icon: "fa fa-ban", status: GradingCriterion_Grade.NONE, style: "secondary", onClick: handleSetAllGrade(GradingCriterion_Grade.NONE) },
        { icon: "fa fa-times", status: GradingCriterion_Grade.FAILED, style: "danger", onClick: handleSetAllGrade(GradingCriterion_Grade.FAILED) },
    ]

    const StatusButtons = buttons.map((button) => {
        // check if all criteria have the same grade as button.status
        const all = review.currentReview?.gradingBenchmarks.every(bm =>
            bm.criteria.every(c => c.grade === button.status))

        // if all criteria have the same grade, use solid button style, else outline
        const style = all ? button.style : `outline-${button.style}`

        return (
            <div role="button" aria-hidden="true" key={button.icon} className={`col btn-xs btn-${style} mr-2 border`} onClick={() => button.onClick()}>
                <i className={button.icon} />
            </div>
        )
    })

    return <div className="btn-group">{StatusButtons}</div>
}

export default GradeAllCriteria
