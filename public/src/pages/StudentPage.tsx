import { Route, Routes, useLocation } from "react-router"
import { CourseLinks } from "../components/CourseLinks"
import Lab from "../components/Lab"
import CourseLabs from "../components/student/CourseLabs"
import SubmissionGuide from "../components/student/SubmissionGuide"
import { useBackspaceNavigation } from "../hooks/useBackspaceNavigation"
import { useCourseID } from "../hooks/useCourseID"
import GroupPage from "./GroupPage"


const StudentPage = () => {
    const courseID = useCourseID()
    const location = useLocation()
    const root = `/course/${courseID}`

    // Enable Backspace keyboard shortcut to navigate back to root
    useBackspaceNavigation(root)

    return (
        <>
            <div hidden={location.pathname !== root}>
                <CourseLinks />
                <CourseLabs />
            </div>
            <Routes>
                <Route path="/group" element={<GroupPage />} />
                <Route path="/submission-guide" element={<SubmissionGuide />} />
                <Route path="/lab/:lab" element={<Lab />} />
                <Route path="/group-lab/:lab" element={<Lab />} />
            </Routes>
        </>
    )
}

export default StudentPage
