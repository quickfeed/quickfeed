import React from "react"
import { Route, Switch, useHistory } from "react-router"
import { getCourseID } from "../Helpers"
import CourseBanner from "../components/CourseBanner"
import CourseLabs from "../components/CourseLabs"
import CourseUtilityLinks from "../components/CourseUtilityLinks"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"
import Alert from "../components/Alert"
import GroupForm from "../components/group/GroupForm"
import GroupComponent from "../components/group/Group"


const StudentPage = (): JSX.Element => {
    const courseID = getCourseID()
    const history = useHistory()
    const root = `/course/${courseID}`

    return (
        <>
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
                <Route path="/course/:id/group" exact component={GroupComponent} />
                <Route path="/course/:id/group/create" exact component={GroupForm} />
                <Route path="/course/:id/:lab" exact component={Lab} />
            </Switch>
        </>
    )
}

export default StudentPage
