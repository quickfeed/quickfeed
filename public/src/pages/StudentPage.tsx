import React from "react"
import { Route, Switch, useHistory } from "react-router"
import CourseLabs from "../components/student/CourseLabs"
import CourseLinks from "../components/CourseLinks"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alerts from "../components/alerts/Alerts"
import { useCourseID } from "../hooks/useCourseID"


const StudentPage = () => {
    const courseID = useCourseID()
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
                <CourseLinks />
            </div>
            <Switch>
                <Route path="/course/:id/group" exact component={GroupPage} />
                <Route path="/course/:id/lab/:lab" exact component={Lab} />
                <Route path="/course/:id/group-lab/:lab" exact component={Lab} />
            </Switch>
        </div>
    )
}

export default StudentPage
