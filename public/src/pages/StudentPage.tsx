import React, { useEffect } from "react"
import { Route, Switch, useHistory } from "react-router"
import { getCourseID } from "../Helpers"
import { useOvermind } from "../overmind"
import CourseBanner from "../components/CourseBanner"
import { CourseLabs } from "../components/CourseLabs"
import CourseUtilityLinks from "../components/CourseUtilityLinks"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import RedirectButton from "../components/RedirectButton"

/* */
const StudentPage = () => {
    const { state } = useOvermind()
    const courseID = getCourseID()
    const history = useHistory()
    const root = `/course/${courseID}`

    return (
        <div className="box">
            <RedirectButton to={root}></RedirectButton>
            <CourseBanner enrollment={state.enrollmentsByCourseId[courseID]} />   
            <div className="row" hidden={history.location.pathname != root}>
                <div className="col-md-9" >
                    <CourseLabs courseID={courseID}/>
                </div>
                <CourseUtilityLinks courseID={courseID} />
            </div>
            <Switch>
                <Route path="/course/:id/group" exact component={GroupPage} />
                <Route path="/course/:id/:lab" exact component={Lab} />
            </Switch>
        </div>
    )
}

export default StudentPage