import React, { useState } from "react"
import { Course } from "../../../proto/qf/types_pb"
import { useAppState } from "../../overmind"
import DynamicTable, { Row } from "../DynamicTable"
import CourseForm from "../forms/CourseForm"


const EditCourse = () => {
    const state = useAppState()
    const [course, setCourse] = useState<Course>()

    const courses = state.courses.map(c => {
        const selected = course?.ID === c.ID
        const data: Row = []
        const badge = selected ? "badge badge-danger" : "badge badge-primary"
        const buttonText = selected ? "Cancel" : "Edit"
        data.push(
            c.name, c.code, c.tag,
            c.year.toString(), c.slipDays.toString(),

            <button className={`clickable ${badge}`} onClick={() => setCourse(selected ? undefined : c)}>
                {buttonText}
            </button>
        )
        return data
    })

    return (
        <div className={"box"}>
            <DynamicTable header={["Course", "Code", "Tag", "Year", "Slipdays", "Edit"]} data={courses} />
            {course ? <CourseForm courseToEdit={course} /> : null}
        </div>
    )
}

export default EditCourse
