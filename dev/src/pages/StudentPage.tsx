import React, { useEffect } from "react"
import { Route, Switch, useHistory } from "react-router"
import { getCourseID } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import CourseBanner from "../components/CourseBanner"
import { CourseLabs } from "../components/CourseLabs"
import CourseUtilityLinks from "../components/CourseUtilityLinks"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alert from "../components/Alert"

/* */
const StudentPage = (): JSX.Element => {
    const state = useAppState()
    const actions = useActions()
    const courseID = getCourseID()
    const history = useHistory()
    const root = `/course/${courseID}`

    useEffect(() => {
        actions.setSelectedEnrollment(state.self.getId())
    })

    return (
        <div>
            <RedirectButton to={root}></RedirectButton>
            <CourseBanner />
            <Alert /> 
            <div className="row" hidden={history.location.pathname != root}>
                <div className="col-md-9" >
                    <CourseLabs />
                </div>
                <CourseUtilityLinks />
            </div>
            <Switch>
                <Route path="/course/:id/group" exact component={GroupPage} />
                <Route path="/course/:id/:lab" exact component={Lab} />
            </Switch>
        </div>
    )
}

export default StudentPage