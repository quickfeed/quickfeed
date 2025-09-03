import React, { useCallback } from "react"
import { Route, Routes, useLocation } from "react-router"
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
import AssignmentFeedbackView from "../components/teacher/AssignmentFeedbackView"

const ReviewResults = () => <Results review />
const RegularResults = () => <Results review={false} />

/* TeacherPage enables routes to be accessed by the teacher only, and displays an overview of the different features available to the teacher. */
const TeacherPage = () => {
    const state = useAppState()
    const actions = useActions().global
    const courseID = useCourseID()
    const location = useLocation()
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
    const feedback = { title: "View Assignment Feedback", text: "View feedback provided by students on assignments.", buttonText: "Feedback", to: `${root}/feedback` }

    return (
        <div className="box">
            <RedirectButton to={root} />
            <Alerts />
            <div className="row" hidden={location.pathname !== root}>
                {courseHasManualGrading && <Card {...review} />}
                <Card {...results} />
                <Card {...groups} />
                <Card {...members} />
                <Card {...assignments} />
                <Card {...updateAssignments} />
                <Card {...feedback} />
            </div>
            <Routes>
                <Route path={"/groups"} element={<GroupPage />} />
                <Route path={"/members"} element={<Members />} />
                <Route path={"/review"} element={<ReviewResults />} />
                <Route path={"/results"} element={<RegularResults />} />
                <Route path={"/assignments"} element={<Assignments />} />
                <Route path={"/feedback"} element={<AssignmentFeedbackView />} />
                <Route path={"/feedback/:assignmentID"} element={<AssignmentFeedbackView />} />
            </Routes>
        </div>
    )
}

export default TeacherPage
