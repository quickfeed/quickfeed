import React, { useCallback, useState } from "react"
import { useActions } from "../../overmind"
import { Course, CourseSchema } from "../../../proto/qf/types_pb"
import FormInput from "./FormInput"
import { useNavigate } from "react-router"
import { clone } from "@bufbuild/protobuf"



// TODO: There are currently issues with navigating a new course without refreshing the page to trigger a state reload.
// TODO: Plenty of required fields are undefined across multiple components.

/** CourseForm is used to create a new course or edit an existing course.
 *  If `editCourse` is provided, the existing course will be modified.
 *  If no course is provided, a new course will be created. */
const CourseForm = ({ courseToEdit }: { courseToEdit: Course }) => {
    const actions = useActions().global
    const navigate = useNavigate()

    // Local state containing the course to be created or edited (if any)
    const [course, setCourse] = useState(clone(CourseSchema, courseToEdit))

    const handleChange = useCallback((event: React.FormEvent<HTMLInputElement>) => {
        const { name, value } = event.currentTarget
        switch (name) {
            case "courseName":
                course.name = value
                break
            case "courseTag":
                course.tag = value
                break
            case "courseCode":
                course.code = value
                break
            case "courseYear":
                course.year = Number(value)
                break
            case "slipDays":
                course.slipDays = Number(value)
                break
        }
        setCourse(course)
    }, [course])

    // Creates a new course if no course is being edited, otherwise updates the existing course
    const submitHandler = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        await actions.editCourse({ course })
        navigate(`/course/${course.ID}`)
    }

    return (
        <div className="container">
            <form className="form-group" onSubmit={async e => await submitHandler(e)}>
                <div className="row">
                    <FormInput prepend="Name"
                        name="courseName"
                        placeholder={"Course Name"}
                        defaultValue={course.name}
                        onChange={handleChange}
                    />
                </div>
                <div className="row">
                    <FormInput
                        prepend="Code"
                        name="courseCode"
                        placeholder={"(ex. DAT320)"}
                        defaultValue={course.code}
                        onChange={handleChange}
                    />
                    <FormInput
                        prepend="Tag"
                        name="courseTag"
                        placeholder={"(ex. Fall / Spring)"}
                        defaultValue={course.tag}
                        onChange={handleChange}
                    />
                </div>
                <div className="row">
                    <FormInput
                        prepend="Slip days"
                        name="slipDays"
                        placeholder={"(ex. 7)"}
                        defaultValue={course.slipDays.toString()}
                        onChange={handleChange}
                        type="number"
                    />
                    <FormInput
                        prepend="Year"
                        name="courseYear"
                        placeholder={"(ex. 2021)"}
                        defaultValue={course.year.toString()}
                        onChange={handleChange}
                        type="number"
                    />
                </div>
                <input className="btn btn-primary" type="submit" value={"Save"} />
            </form>
        </div>
    )
}

export default CourseForm
