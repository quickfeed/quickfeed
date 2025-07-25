import React from "react"
import { useAppState } from "../overmind"
import { Enrollment_UserStatus, EnrollmentSchema } from "../../proto/qf/types_pb"
import CourseCard from "./CourseCard"
import Button, { ButtonType } from "./admin/Button"
import { useNavigate } from "react-router"
import { Color, isVisible } from "../Helpers"
import { create } from "@bufbuild/protobuf"

// If home is set to true, display only favorite courses. Otherwise, display all courses.
// Can be used on dashboard to let the user choose which courses to display based on favorites.
interface overview {
    home: boolean
}

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
                            type={ButtonType.BUTTON}
                            className="mr-3"
                            onClick={() => navigate("/admin/create")}
                        />
                        <Button
                            text="Manage users"
                            color={Color.BLUE}
                            type={ButtonType.BUTTON}
                            onClick={() => navigate("/admin/manage")}
                        />
                    </div>
                    : null}
            </div>
        )
    }

    // Push to separate arrays for layout purposes. Favorite - Student - Teacher - Pending
    const courses = () => {
        const favorite: React.JSX.Element[] = []
        const student: React.JSX.Element[] = []
        const teacher: React.JSX.Element[] = []
        const pending: React.JSX.Element[] = []
        const availableCourses: React.JSX.Element[] = []
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
            } else {
                availableCourses.push(
                    <CourseCard key={course.ID.toString()} course={course} enrollment={create(EnrollmentSchema)} />
                )
            }
        })

        if (overview.home) {
            // Render only favorite courses.
            return (
                <>
                    {favorite.length > 0 &&
                        <div className="container-fluid">
                            <div className="card-deck course-card-row favorite-row">
                                {favorite}
                            </div>
                        </div>
                    }
                </>
            )
        }

        return (
            <div className="box container-fluid">
                {favorite.length > 0 &&
                    <div className="container-fluid">
                        <h2>Favorites</h2>
                        <div className="card-deck course-card-row favorite-row">
                            {favorite}
                        </div>
                    </div>
                }

                {(student.length > 0 || teacher.length > 0) &&
                    <div className="container-fluid myCourses">
                        <h2>My Courses</h2>
                        <div className="card-deck course-card-row">
                            {teacher}
                            {student}
                        </div>
                    </div>
                }
                {pending.length > 0 &&
                    <div className="container-fluid">
                        {(student.length === 0 && teacher.length === 0) &&
                            <h2>My Courses</h2>
                        }
                        <div className="card-deck">
                            {pending}
                        </div>
                    </div>
                }

                {availableCourses.length > 0 &&
                    <>
                        <h2>Available Courses</h2>
                        <div className="card-deck course-card-row">
                            {availableCourses}
                        </div>
                    </>
                }
            </div>
        )
    }
    return courses()

}

export default Courses
