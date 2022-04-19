import React from "react"
import { useAppState } from "../overmind"
import { Enrollment } from "../../proto/ag/ag_pb"
import CourseCard from "./CourseCard"
import Button, { ButtonType } from "./admin/Button"
import { useHistory } from "react-router"
import { Color, isVisible } from "../Helpers"

// If home is set to true, display only favorite courses. Otherwise, display all courses.
// Can be used on dashboard to let the user choose which courses to display based on favorites.
interface overview {
    home: boolean
}

/** This component lists the user's courses and courses available for enrollment. */
const Courses = (overview: overview): JSX.Element => {
    const state = useAppState()
    const history = useHistory()

    // Notify user if there are no courses (should only ever happen with a fresh database on backend)
    // Display shortcut buttons for admins to create new course or managing (promoting) users
    if (state.courses.length == 0) {
        return (
            <div className="container centered">
                <h3>There are currently no available courses.</h3>
                {state.self.getIsadmin() ?
                    <>
                        <div>
                            <Button classname="mr-3" text="Go to course creation" color={Color.GREEN} type={ButtonType.BUTTON} onclick={() => history.push("/admin/create")} />
                            <Button text="Manage users" color={Color.BLUE} type={ButtonType.BUTTON} onclick={() => history.push("/admin/manage")} />
                        </div>
                    </>
                    : null}
            </div>
        )
    }

    // Push to separate arrays for layout purposes. Favorite - Student - Teacher - Pending
    const courses = () => {
        const favorite: JSX.Element[] = []
        const student: JSX.Element[] = []
        const teacher: JSX.Element[] = []
        const pending: JSX.Element[] = []
        const availableCourses: JSX.Element[] = []
        state.courses.map(course => {
            const enrol = state.enrollmentsByCourseID[course.getId()]
            if (enrol) {
                const courseCard = <CourseCard key={course.getId()} course={course} enrollment={enrol} />
                if (isVisible(enrol)) {
                    favorite.push(courseCard)
                } else {
                    switch (enrol.getStatus()) {
                        case Enrollment.UserStatus.PENDING:
                            pending.push(courseCard)
                            break
                        case Enrollment.UserStatus.STUDENT:
                            student.push(courseCard)
                            break
                        case Enrollment.UserStatus.TEACHER:
                            teacher.push(courseCard)
                            break
                    }
                }
            } else {
                availableCourses.push(
                    <CourseCard key={course.getId()} course={course} enrollment={new Enrollment} />
                )
            }
        })

        if (overview.home) {
            // Render only favorited courses.
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
                    <div className="container-fluid">
                        <h2>My Courses</h2>
                        <div className="card-deck course-card-row">
                            {teacher}
                            {student}
                        </div>
                    </div>
                }
                {pending.length > 0 &&
                    <div className="container-fluid">
                        {(student.length == 0 && teacher.length == 0) &&
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
