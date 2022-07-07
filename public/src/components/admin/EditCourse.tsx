import React, { useState } from "react"
import { Course } from "../../../proto/qf/types/types_pb"
import { useAppState } from "../../overmind"
import DynamicTable, { Row } from "../DynamicTable"
import CourseForm from "../forms/CourseForm"


const EditCourse = (): JSX.Element => {
    const state = useAppState()
    const [course, setCourse] = useState<Course.AsObject>()

    const courses = state.courses.map(c => {
        const selected = course?.id === c.id
        const data: Row = []
        data.push(c.name)
        data.push(c.code)
        data.push(c.tag)
        data.push(c.year.toString())
        data.push(c.slipdays.toString())
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
            {course ? <CourseForm editCourse={course} /> : null}
        </div>
    )
}

export default EditCourse
