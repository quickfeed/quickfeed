import { useCourseID } from "../../hooks/useCourseID"
import { useAppState } from "../../overmind"
import CourseForm from "../forms/CourseForm"

/** EditCourse lets a teacher edit the details of the course they are currently viewing.
 *  It is reached from the teacher page at /course/:id/edit. */
const EditCourse = () => {
    const state = useAppState()
    const courseID = useCourseID()
    const course = state.courses.find(c => c.ID === courseID)

    if (!course) {
        return <p className="text-base-content/60">Course not found.</p>
    }

    return <CourseForm courseToEdit={course} />
}

export default EditCourse
