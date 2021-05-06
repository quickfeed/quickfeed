import React, { useEffect } from "react"
import { Redirect, RouteComponentProps } from "react-router"
import { Switch, Route } from "react-router-dom"

import { useOvermind } from "../overmind"


import CourseOverview from "./CourseOverview"
import Group from "./Group"
import Groups from "./Groups"
import Lab from "./Lab"
import Members from "./Members"
import Review from "./Review"



interface MatchProps {
    id: string
}


const CoursePage = (props: RouteComponentProps<MatchProps>) => {
    const { state, actions: {setActiveCourse} } = useOvermind()
    let courseID = Number(props.match.params.id)
    useEffect(() => {
            setActiveCourse(courseID)
    }, [props])

    if (state.enrollmentsByCourseId[courseID]){
        return (
        <Switch>
            <Route path="/course/:id" exact component={CourseOverview} />
            <Route path="/course/:id/group" exact component={Group} />
            <Route path="/course/:id/members" exact component={Members} />
            <Route path="/course/:id/review" exact component={Review} />
            <Route path="/course/:id/groups" exact component={Groups} />
            <Route path="/course/:id/:lab" exact component={Lab} />
        </Switch>
        )
    }
    else {
        return (<Redirect to={"/"}></Redirect>)
    }
}

export default CoursePage