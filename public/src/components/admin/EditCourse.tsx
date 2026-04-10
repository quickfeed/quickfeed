import React, { useState } from "react"
import { Course } from "../../../proto/qf/types_pb"
import { useAppState } from "../../overmind"
import DynamicTable from "../DynamicTable"
import CourseForm from "../forms/CourseForm"


const EditCourse = () => {
    const state = useAppState()
    const [course, setCourse] = useState<Course>()

    const courses = state.courses.map(c => {
        const selected = course?.ID === c.ID
        const badge = selected ? "badge badge-error" : "badge badge-primary"
        const buttonText = selected ? "Cancel" : "Edit"
        return [
            c.name,
            c.code,
            c.tag,
            c.year.toString(),
            c.slipDays.toString(),
            <button key={c.ID} className={`clickable ${badge}`} onClick={() => setCourse(selected ? undefined : c)}>
                {buttonText}
            </button>
        ]
    })

    return (
        <div className={"box"}>
            <DynamicTable header={["Course", "Code", "Tag", "Year", "Slipdays", "Edit"]} data={courses} />
            {course ? <CourseForm courseToEdit={course} /> : null}
        </div>
    )
}

export default EditCourse
