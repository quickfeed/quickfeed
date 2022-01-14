import React from "react"
import { GradingCriterion } from "../../../proto/ag/ag_pb"
import { useActions, useAppState } from "../../overmind"

const CriteriaStatus = ({ criterion }: { criterion: GradingCriterion }): JSX.Element => {
    const { setGrade } = useActions().review
    const { isTeacher } = useAppState()

    if (!isTeacher) {
        return <></>
    }

    const buttons: { icon: string, status: GradingCriterion.Grade, style: string, onClick: () => void }[] = [
        { icon: "fa fa-check", status: GradingCriterion.Grade.PASSED, style: "success", onClick: () => setGrade({ criterion: criterion, grade: GradingCriterion.Grade.PASSED }) },
        { icon: "fa fa-ban", status: GradingCriterion.Grade.NONE, style: "secondary", onClick: () => setGrade({ criterion: criterion, grade: GradingCriterion.Grade.NONE }) },
        { icon: "fa fa-times", status: GradingCriterion.Grade.FAILED, style: "danger", onClick: () => setGrade({ criterion: criterion, grade: GradingCriterion.Grade.FAILED }) },
    ]

    const StatusButtons = buttons.map((button, index) => {
        const style = criterion.getGrade() === button.status ? `col btn-xs btn-${button.style} mr-2 border` : `col btn-xs btn-outline-${button.style} mr-2 border`
        // TODO: Perhaps refactor button into a separate general component to enable reuse
        return (
            <div key={index} className={style} onClick={() => button.onClick()}>
                <i className={button.icon}></i>
            </div>
        )
    })

    return <div className="btn-group">{StatusButtons}</div>
}

export default CriteriaStatus
