import React from "react"
import { Route, Routes, useLocation } from "react-router"
import CourseLabs from "../components/student/CourseLabs"
import GroupPage from "./GroupPage"
import Lab from "../components/Lab"
import { useCourseID } from "../hooks/useCourseID"
import { RepositoryCards } from "../components/student/RepositoryCards"
import { useBackspaceNavigation } from "../hooks/useBackspaceNavigation"


const StudentPage = () => {
    const courseID = useCourseID()
    const location = useLocation()
    const root = `/course/${courseID}`

    // Enable Backspace keyboard shortcut to navigate back to root
    useBackspaceNavigation(root)

    return (
        <>
            <div hidden={location.pathname !== root}>
                {/* Compact top bar with repository links and group navigation */}
                <div className="flex flex-wrap items-center gap-x-6 gap-y-2 mt-3 mb-4 px-3 py-2 bg-base-200 rounded-lg">
                    <RepositoryCards />
                </div>
                <CourseLabs />
            </div>
            <Routes>
                <Route path="/group" element={<GroupPage />} />
                <Route path="/lab/:lab" element={<Lab />} />
                <Route path="/group-lab/:lab" element={<Lab />} />
            </Routes>
        </>
    )
}

export default StudentPage
