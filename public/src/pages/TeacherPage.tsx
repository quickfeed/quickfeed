import React, { useCallback } from "react"
import { Route, Switch, useHistory } from "react-router"
import { Color, isManuallyGraded } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import Card from "../components/Card"
import GroupPage from "./GroupPage"
import Members from "../components/Members"
import RedirectButton from "../components/RedirectButton"
import Results from "../components/Results"
import Assignments from "../components/teacher/Assignments"
import Alerts from "../components/alerts/Alerts"
import { useCourseID } from "../hooks/useCourseID"

const ReviewResults = () => <Results review />
const RegularResults = () => <Results review={false} />

/* TeacherPage enables routes to be accessed by the teacher only, and displays an overview of the different features available to the teacher. */
const TeacherPage = () => {
    const state = useAppState()
    const actions = useActions()
    const courseID = useCourseID()
    const history = useHistory()
    const root = `/course/${courseID}`
    const courseHasManualGrading = state.assignments[courseID.toString()]?.some(assignment => isManuallyGraded(assignment.reviewers))

    const members = {
        title: "View Members",
        notification: state.pendingEnrollments.length > 0 ? { color: Color.YELLOW, text: "Pending enrollments" } : undefined,
        text: "View all students, and approve new enrollments.",
        buttonText: "Members", to: `${root}/members`
    }
    const groups = {
        title: "Manage Groups",
        notification: state.pendingGroups.length > 0 ? { color: Color.YELLOW, text: "Pending groups" } : undefined,
        text: "View, edit or delete course groups.",
        buttonText: "Groups", to: `${root}/groups`
    }
    const results = { title: "View results", text: "View results for all students in the course.", buttonText: "Results", to: `${root}/results` }
    const assignments = { title: "Manage Assignments", text: "View and edit assignments.", buttonText: "Assignments", to: `${root}/assignments` }
    const handleUpdateAssignments = useCallback(() => actions.updateAssignments(courseID), [actions, courseID])
    const updateAssignments = {
        title: "Update Course Assignments",
        text: "Fetch assignments from GitHub.",
        buttonText: "Update Assignments",
        onclick: handleUpdateAssignments
    }
    const review = { title: "Review Assignments", text: "Review assignments for students.", buttonText: "Review", to: `${root}/review` }

    return (
        <div className="box">
            <RedirectButton to={root} />
            <Alerts />
            <div className="row" hidden={history.location.pathname != root}>
                {courseHasManualGrading && <Card {...review} />}
                <Card {...results} />
                <Card {...groups} />
                <Card {...members} />
                <Card {...assignments} />
                <Card {...updateAssignments} />
            </div>
            <Switch>
                <Route path={`/course/:id/groups`} exact component={GroupPage} />
                <Route path={"/course/:id/members"} component={Members} />
                <Route path={"/course/:id/review"} component={ReviewResults} />
                <Route path={"/course/:id/results"} component={RegularResults} />
                <Route path={"/course/:id/assignments"} component={Assignments} />
            </Switch>
        </div>
    )
}

export default TeacherPage
