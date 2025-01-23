import React, { useState } from "react"
import { Course } from "../../../proto/qf/types_pb"
import { useAppState } from "../../overmind"
import DynamicTable, { Row } from "../DynamicTable"
import CourseForm from "../forms/CourseForm"


const EditCourse = (): JSX.Element => {
    const state = useAppState()
    const [course, setCourse] = useState<Course>()

    const courses = state.courses.map(c => {
        const selected = course?.ID === c.ID
        const data: Row = []
        data.push(c.name)
        data.push(c.code)
        data.push(c.tag)
        data.push(c.year.toString())
        data.push(c.slipDays.toString())
        data.push(
            <span className={selected ? "badge badge-danger clickable" : "badge badge-primary clickable"}
                onClick={() => { selected ? setCourse(undefined) : setCourse(c) }}>
                {selected ? "Cancel" : "Edit"}
            </span>
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
