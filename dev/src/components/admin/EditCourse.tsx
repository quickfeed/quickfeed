import React, { useState } from "react"
import { Course } from "../../../proto/ag/ag_pb"
import { useAppState } from "../../overmind"
import DynamicTable, { CellElement } from "../DynamicTable"
import CourseForm from "../forms/CourseForm"

const EditCourse = (): JSX.Element => {
    const state  = useAppState()
    const [course, setCourse] = useState<Course>()

    const courses = state.courses.map(c => {
        const selected = course?.getId() === c.getId()
        const data: (string | CellElement)[] = []
        data.push(c.getName())
        data.push(c.getCode())
        data.push(c.getTag())
        data.push(c.getYear().toString())
        data.push(c.getSlipdays().toString())
        data.push({value: selected ? "Cancel" : "Edit", className: selected ? "btn btn-danger" : "btn btn-primary", onClick: selected ? () => setCourse(undefined) : () => setCourse(c)})
        return data
    })

    return (
        <div className={"box"}>
            <DynamicTable header={["Course", "Code", "Tag", "Year", "Slipdays", "Edit"]} data={courses} />
            {course ? <CourseForm editCourse={course} /> : null}
        </div>
    )
}

export default EditCourse