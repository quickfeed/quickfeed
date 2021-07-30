import React, { useEffect, useState } from "react"
import { useHistory } from "react-router"
import { Course } from "../../../proto/ag/ag_pb"
import { useAppState } from "../../overmind"
import CourseForm from "../forms/CourseForm"

/** TODO:  */
const EditCourse = () => {
    const state  = useAppState()
    const history = useHistory()
    const [course, setCourse] = useState<Course>()
    useEffect(() => {
    }, [course, setCourse])


    const courses = state.courses.map(course => {
        return (
            <tr>
                <th colSpan={2}>{course.getName()}</th>
                <td>{course.getCode()}</td>
                <td>{course.getTag()}</td>
                <td>{course.getYear()}</td>
                <td>{course.getSlipdays()}</td>
                <td><div className={"btn btn-primary"} onClick={() => setCourse(course)}>Edit</div></td>
            </tr>
        )
    })

    return (
        <div className={"box"}>
            <table className="table table-curved table-striped">
                <thead className={"thead-dark"}>
                <th colSpan={2}>Course</th>
                <th>Code</th>
                <th>Tag</th>
                <th>Year</th>
                <th>Slipdays</th>
                <th>Edit</th>
                </thead>
                <tbody>
                {courses}
                </tbody>

            </table>
            {course ? <CourseForm editCourse={course} /> : null}
        </div>
    )
}

export default EditCourse