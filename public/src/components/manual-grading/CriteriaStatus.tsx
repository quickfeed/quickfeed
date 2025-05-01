import React from "react"
import { GradingCriterion, GradingCriterion_Grade } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"


const CriteriaStatus = ({ criterion }: { criterion: GradingCriterion }) => {
    const { setGrade } = useActions().review
    const { isTeacher } = useAppState()

    const handleSetGrade = (grade: GradingCriterion_Grade) => () => setGrade({ criterion, grade })

    if (!isTeacher) {
        return null
    }

    const buttons: { icon: string, status: GradingCriterion_Grade, style: string, onClick: () => void }[] = [
        { icon: "fa fa-check", status: GradingCriterion_Grade.PASSED, style: "success", onClick: handleSetGrade(GradingCriterion_Grade.PASSED) },
        { icon: "fa fa-ban", status: GradingCriterion_Grade.NONE, style: "secondary", onClick: handleSetGrade(GradingCriterion_Grade.NONE) },
        { icon: "fa fa-times", status: GradingCriterion_Grade.FAILED, style: "danger", onClick: handleSetGrade(GradingCriterion_Grade.FAILED) },
    ]

    const StatusButtons = buttons.map((button) => {
        const style = criterion.grade === button.status ? button.style : `outline-${button.style}`
        // TODO: Perhaps refactor button into a separate general component to enable reuse
        return (
            <div role="button" aria-hidden="true" key={button.icon} className={`col btn-xs btn-${style} mr-2 border`} onClick={() => button.onClick()}>
                <i className={button.icon} />
            </div>
        )
    })

    return <div className="btn-group">{StatusButtons}</div>
}

export default CriteriaStatus
