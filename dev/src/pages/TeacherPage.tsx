import React from "react"
import { Route, Switch, useHistory } from "react-router"
import { Color, getCourseID, isManuallyGraded } from "../Helpers"
import { useAppState, useGrpc } from "../overmind"
import Card from "../components/Card"
import CourseBanner from "../components/CourseBanner"
import GroupPage from "./GroupPage"
import Members from "../components/Members"
import RedirectButton from "../components/RedirectButton"
import Results from "../components/Results"
import ReviewPage from "../components/ReviewPage"
import Assignments from "../components/teacher/Assignments"
import Alert from "../components/Alert"


/* TeacherPage enables routes to be accessed by the teacher only, and displays an overview of the different features available to the teacher. */
const TeacherPage = (): JSX.Element => {
    const state = useAppState()
    const courseID = getCourseID()
    const grpc = useGrpc().grpcMan
    const history = useHistory()
    const root = `/course/${courseID}`
    const courseHasManualGrading = state.assignments[courseID].some(assignment => isManuallyGraded(assignment))

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
    const updateAssignments = { title: "Update Course Assignments", text: "Fetch assignments from GitHub.", buttonText: "Update Assignments", onclick: () => grpc.updateAssignments(courseID) }
    const review = { title: "Review Assignments", text: "Review assignments for students.", buttonText: "Review", to: `${root}/review` }

    return (
        <div>
            <RedirectButton to={root}></RedirectButton>
            <CourseBanner />
            <Alert />
            <div className="row" hidden={history.location.pathname != root}>
                {courseHasManualGrading && <Card {...review} />}
                <Card {...results}></Card>
                <Card {...groups}></Card>
                <Card {...members}></Card>
                <Card {...assignments}></Card>
                <Card {...updateAssignments}></Card>
            </div>
            <Switch>
                <Route path={`/course/:id/groups`} exact component={GroupPage}></Route>
                <Route path={"/course/:id/members"} component={Members}></Route>
                <Route path={"/course/:id/review"} component={ReviewPage}></Route>
                <Route path={"/course/:id/results"} component={Results}></Route>
                <Route path={"/course/:id/assignments"} component={Assignments}></Route>
            </Switch>
        </div>
    )
}

export default TeacherPage
