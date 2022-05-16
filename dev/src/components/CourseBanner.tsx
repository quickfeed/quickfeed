import React from "react"
import { getCourseID, hasTeacher, isTeacher, isVisible } from "../Helpers"
import { useActions, useAppState } from "../overmind"


// TODO: Maybe add route specific information, ex. if user is viewing a lab, show that in the banner. Could use state in components to display.
// TODO(jostein): This information could possibly be shown in the navbar.
const CourseBanner = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()

    const enrollment = state.enrollmentsByCourseID[getCourseID()]
    const style = isVisible(enrollment) ? "fa fa-star" : "fa fa-star-o"
    return (
        <div className="banner jumbotron">
            <div className="centerblock container">
                <h1>{enrollment.getCourse()?.getName()}
                    <i className={style} style={{ "paddingLeft": "20px" }}
                        onClick={() => actions.setEnrollmentState(enrollment)} />
                </h1>
                {hasTeacher(state.status[enrollment.getCourseid()]) &&
                    <span className="clickable" onClick={() => actions.changeView(enrollment.getCourseid())}>
                        {isTeacher(enrollment) ? "Switch to Student View" : "Switch to Teacher View"}
                    </span>
                }
            </div>
        </div>
    )
}

export default CourseBanner
