import React from "react"
import { Route, Switch, useHistory } from "react-router"
import { getCourseID } from "../Helpers"
import CourseLabs from "../components/CourseLabs"
import CourseUtilityLinks from "../components/CourseUtilityLinks"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alerts from "../components/alerts/Alerts"


const StudentPage = (): JSX.Element => {
    const courseID = getCourseID()
    const history = useHistory()
    const root = `/course/${courseID}`

    return (
        <div className="box">
            <RedirectButton to={root} />
            <Alerts />
            <div className="row" hidden={history.location.pathname !== root}>
                <div className="col-md-9" >
                    <CourseLabs />
                </div>
                <CourseUtilityLinks />
            </div>
            <Switch>
                <Route path="/course/:id/group" exact component={GroupPage} />
                <Route path="/course/:id/lab/:lab" exact component={Lab} />
            </Switch>
        </div>
    )
}

export default StudentPage
