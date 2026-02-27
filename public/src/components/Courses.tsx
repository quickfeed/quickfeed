import React, { ComponentProps } from "react"
import Collapsible from "./Collapsible"
import { useAppState } from "../overmind"
import { Course, Enrollment_UserStatus, EnrollmentSchema } from "../../proto/qf/types_pb"
import CourseCard from "./CourseCard"
import Button from "./admin/Button"
import { useNavigate } from "react-router"
import { Color, isVisible } from "../Helpers"
import { create } from "@bufbuild/protobuf"

// If home is set to true, display only favorite courses. Otherwise, display all courses.
// Can be used on dashboard to let the user choose which courses to display based on favorites.
interface overview {
    home: boolean
}

// Type for a course card element
type CourseCardElement = React.ReactElement<ComponentProps<typeof CourseCard>>

// Reusable component for course sections with icon, title, and grid
const CourseSection = ({ icon, title, children }: { icon: string; title: string; children: React.ReactNode }) => (
    <div className="mb-10">
        <div className="flex items-center gap-3 mb-6">
            <i className={`fa ${icon} text-primary text-xl`} />
            <h2 className="text-3xl font-bold text-base-content">{title}</h2>
            <div className="flex-grow h-px bg-gradient-to-r from-base-300 to-transparent ml-4" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {children}
        </div>
    </div>
)

/** This component lists the user's courses and courses available for enrollment. */
const Courses = (overview: overview) => {
    const state = useAppState()
    const navigate = useNavigate()

    // Notify user if there are no courses (should only ever happen with a fresh database on backend)
    // Display shortcut buttons for admins to create new course or managing (promoting) users
    if (state.courses.length === 0) {
        return (
            <div className="container centered">
                <h3>There are currently no available courses.</h3>
                {state.self.IsAdmin ?
                    <div>
                        <Button
                            text="Go to course creation"
                            color={Color.GREEN}
                            className="mr-3"
                            onClick={() => navigate("/admin/create")}
                        />
                        <Button
                            text="Manage users"
                            color={Color.BLUE}
                            onClick={() => navigate("/admin/manage")}
                        />
                    </div>
                    : null}
            </div>
        )
    }

    // Push to separate arrays for layout purposes. Favorite - Student - Teacher - Pending
    const favorite: CourseCardElement[] = []
    const student: CourseCardElement[] = []
    const teacher: CourseCardElement[] = []
    const pending: CourseCardElement[] = []
    const availableCourses: CourseCardElement[] = []
    const unavailableCourses: CourseCardElement[] = []
    state.courses.forEach(course => {
        const enrol = state.enrollmentsByCourseID[course.ID.toString()]
        if (enrol) {
            const courseCard = <CourseCard key={course.ID.toString()} course={course} enrollment={enrol} />
            if (isVisible(enrol)) {
                favorite.push(courseCard)
            } else {
                switch (enrol.status) {
                    case Enrollment_UserStatus.PENDING:
                        pending.push(courseCard)
                        break
                    case Enrollment_UserStatus.STUDENT:
                        student.push(courseCard)
                        break
                    case Enrollment_UserStatus.TEACHER:
                        teacher.push(courseCard)
                        break
                }
            }
            return
        }

        if (shouldBeUnavailable(course)) {
            unavailableCourses.push(
                <CourseCard key={course.ID.toString()} course={course} enrollment={create(EnrollmentSchema)} unavailable />
            )
            return
        }
        availableCourses.push(
            <CourseCard key={course.ID.toString()} course={course} enrollment={create(EnrollmentSchema)} />
        )

    })

    // sort courses by year and term, most recent first
    const sortByYearTerm = (a: CourseCardElement, b: CourseCardElement) => {
        const courseA = a.props.course
        const courseB = b.props.course
        if (courseA.year !== courseB.year) {
            return courseB.year - courseA.year // Descending order for year
        }

        // Map terms to an order value
        const termOrder: Record<string, number> = {
            Fall: 2,
            Spring: 1,
        }
        // tag is used to represent term, e.g. "Spring" < "Fall"
        // fall should come before spring in the same year as that is
        // the more recent term
        return termOrder[courseB.tag] - termOrder[courseA.tag]
    }

    if (overview.home) {
        // Render only favorite courses.
        return (
            favorite.length > 0 &&
            <div className="container">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {favorite}
                </div>
            </div>
        )
    }

    return (
        <div className="container mx-auto px-4 mb-5 mt-4">
            {favorite.length > 0 && (
                <CourseSection icon="fa-star" title="Favorites">
                    {favorite}
                </CourseSection>
            )}

            {(student.length > 0 || teacher.length > 0) && (
                <CourseSection icon="fa-graduation-cap" title="My Courses">
                    {teacher.sort(sortByYearTerm)}
                    {student.sort(sortByYearTerm)}
                </CourseSection>
            )}

            {pending.length > 0 && (student.length === 0 && teacher.length === 0) && (
                <CourseSection icon="fa-graduation-cap" title="My Courses">
                    {pending}
                </CourseSection>
            )}

            {pending.length > 0 && (student.length > 0 || teacher.length > 0) && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 mb-10">
                    {pending}
                </div>
            )}

            {availableCourses.length > 0 && (
                <CourseSection icon="fa-book" title="Available Courses">
                    {availableCourses.sort(sortByYearTerm)}
                </CourseSection>
            )}

            {unavailableCourses.length > 0 && (
                <Collapsible title={`Unavailable Courses (${unavailableCourses.length})`}>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                        {unavailableCourses.sort(sortByYearTerm)}
                    </div>
                </Collapsible>
            )}
        </div>
    )
}

const shouldBeUnavailable = (course: Course): boolean => {
    const now = new Date()
    return now.getFullYear() > course.year
}

export default Courses
