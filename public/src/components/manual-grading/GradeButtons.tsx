import React from "react"
import { GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { Color } from "../../Helpers"
import Button, { ButtonType } from "../admin/Button"

const GRADE_BUTTONS: { icon: string; grade: GradingCriterion_Grade; color: Color }[] = [
    { icon: "fa fa-check", grade: GradingCriterion_Grade.PASSED, color: Color.GREEN },
    { icon: "fa fa-ban", grade: GradingCriterion_Grade.NONE, color: Color.GRAY },
    { icon: "fa fa-times", grade: GradingCriterion_Grade.FAILED, color: Color.RED },
]

interface GradeButtonsProps {
    /** Returns true if the given grade is the currently active one (solid style). */
    isActive: (grade: GradingCriterion_Grade) => boolean
    onClick: (grade: GradingCriterion_Grade) => void
}

const GradeButtons = ({ isActive, onClick }: GradeButtonsProps) => (
    <div className="btn-group">
        {GRADE_BUTTONS.map(({ icon, grade, color }) => (
            <Button
                text=""
                key={icon}
                color={color}
                type={isActive(grade) ? ButtonType.SOLID : ButtonType.GHOST}
                className="btn-md mr-2"
                onClick={() => onClick(grade)}
            >
                <i className={icon} />
            </Button>
        ))}
    </div>
)

export default GradeButtons
