import React from "react"
import { GradingCriterion, GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import { Color } from "../../Helpers"


const CriteriaStatus = ({ criterion }: { criterion: GradingCriterion }) => {
    const { setGrade } = useActions().review
    const { isTeacher } = useAppState()

    const handleSetGrade = (grade: GradingCriterion_Grade) => () => setGrade({ criterion, grade })

    if (!isTeacher) {
        return null
    }

    const buttons: { icon: string, status: GradingCriterion_Grade, style: Color, onClick: () => void }[] = [
        { icon: "fa fa-check", status: GradingCriterion_Grade.PASSED, style: Color.GREEN, onClick: handleSetGrade(GradingCriterion_Grade.PASSED) },
        { icon: "fa fa-ban", status: GradingCriterion_Grade.NONE, style: Color.GRAY, onClick: handleSetGrade(GradingCriterion_Grade.NONE) },
        { icon: "fa fa-times", status: GradingCriterion_Grade.FAILED, style: Color.RED, onClick: handleSetGrade(GradingCriterion_Grade.FAILED) },
    ]

    const StatusButtons = buttons.map((button) => {
        const buttonType = criterion.grade === button.status ? ButtonType.SOLID : ButtonType.GHOST
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

export default CriteriaStatus
