import React, { useEffect } from "react"
import { Redirect } from "react-router"
import { Switch, Route } from "react-router-dom"
import { Enrollment } from "../../proto/ag/ag_pb"
import { getCourseID } from "../Helpers"
import { useOvermind } from "../overmind"
import CourseOverview from "./CourseOverview"
import { GroupPage } from "./GroupPage"
import Groups from "./Groups"
import Lab from "./Lab"
import Members from "./Members"
import Results from "./Results"
import Review from "./Review"
import StatisticsView from "./Statistics"


const CoursePage = () => {
    const { state, actions: {setActiveCourse} } = useOvermind()
    let courseID = getCourseID()
    useEffect(() => {
        setActiveCourse(courseID)
    }, [courseID])

    if (state.enrollmentsByCourseId[courseID] && state.enrollmentsByCourseId[courseID].getStatus() >= Enrollment.UserStatus.STUDENT){
        return (
            <Switch>
                <Route path="/course/:id" exact component={CourseOverview} />
                <Route path="/course/:id/group" exact component={GroupPage} />
                <Route path="/course/:id/members" exact component={Members} />
                <Route path="/course/:id/review" exact component={Review} />
                <Route path="/course/:id/groups" exact component={Groups} />
                <Route path="/course/:id/results" exact component={Results} />
                <Route path="/course/:id/statistics" exact component={StatisticsView} />
                <Route path="/course/:id/:lab" exact component={Lab} />
            </Switch>
        )
    }
    else {
        return (<Redirect to={"/"}></Redirect>)
    }
}

export default CoursePage