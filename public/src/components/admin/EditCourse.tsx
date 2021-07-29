import React from "react"
import { useAppState } from "../../overmind"

/** TODO:  */
const EditCourse = () => {
    const state  = useAppState()

    const courses = state.courses.map(course => {
        return (
            <tr style={{"cursor": "pointer"}}>
                <th colSpan={2}>{course.getName()}</th>
                <td>{course.getCode()}</td>
                <td>{course.getTag()}</td>
                <td>{course.getYear()}</td>
                <td>{course.getSlipdays()}</td>
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
                </thead>
                <tbody>
                {courses}
                </tbody>

            </table>
        </div>
    )
}

export default EditCourse