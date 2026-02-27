import React from "react"
import { GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import DynamicButton from "../DynamicButton"
import Button, { ButtonType } from "../admin/Button"
import { Color } from "../../Helpers"


const GradeAllCriteria = () => {
    const actions = useActions().review
    const { isTeacher, review } = useAppState()

    if (!isTeacher) {
        return null
    }

    const handleSetAllGrade = (grade: GradingCriterion_Grade) => () => {
        actions.setAllGrade(grade)
    }

    const buttons: { icon: string, status: GradingCriterion_Grade, style: Color, onClick: () => void }[] = [
        { icon: "fa fa-check", status: GradingCriterion_Grade.PASSED, style: Color.GREEN, onClick: handleSetAllGrade(GradingCriterion_Grade.PASSED) },
        { icon: "fa fa-ban", status: GradingCriterion_Grade.NONE, style: Color.GRAY, onClick: handleSetAllGrade(GradingCriterion_Grade.NONE) },
        { icon: "fa fa-times", status: GradingCriterion_Grade.FAILED, style: Color.RED, onClick: handleSetAllGrade(GradingCriterion_Grade.FAILED) },
    ]

    const StatusButtons = buttons.map((button) => {
        // check if all criteria have the same grade as button.status
        const all = review.currentReview?.gradingBenchmarks.every(bm =>
            bm.criteria.every(c => c.grade === button.status))

        // if all criteria have the same grade, use solid button style, else outline
        const buttonType = all ? ButtonType.SOLID : ButtonType.GHOST

        return (
            <Button
                text=""
                key={button.icon}
                color={button.style}
                type={buttonType}
                className={`btn-md mr-2`}
                onClick={() => button.onClick()}
            >
                <i className={button.icon} />
            </Button>
        )
    })

    return <div className="btn-group">{StatusButtons}</div>
}

export default GradeAllCriteria
