import { Navigate } from "react-router"
import { hasEnrollment, isEnrolled, isVisible } from "../Helpers"
import { useAppState } from "../overmind"
import Courses from "./Courses"

/* Dashboard for a signed in user. */
const Dashboard = () => {
    const state = useAppState()

    // Users that are not enrolled in any courses are redirected to the course list.
    if (!state.isLoading && !hasEnrollment(state.enrollments)) {
        return <Navigate to="/courses" />
    }

    // With only one favorite course, the dashboard would show a single course card;
    // go straight to that course instead, whether the user is a teacher or a student.
    // The sidebar's "View all courses" link (/courses) remains the way to see every course.
    const favorites = state.enrollments.filter(enrollment => isEnrolled(enrollment) && isVisible(enrollment))
    if (!state.isLoading && favorites.length === 1) {
        return <Navigate to={`/course/${favorites[0].courseID}`} replace />
    }

    return (
        <div className="mt-5">
            <Courses home />
        </div>
    )
}

export default Dashboard
