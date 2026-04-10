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
        <div className="card bg-base-200 shadow-xl max-w-2xl mx-auto">
            <div className="card-body">
                <h2 className="card-title text-2xl mb-6">
                    <i className="fa fa-book mr-2" />
                    Edit Course
                </h2>
                <form className="space-y-6" onSubmit={async e => await submitHandler(e)}>
                    <div className="grid grid-cols-1 gap-4">
                        <FormInput
                            prepend="Name"
                            name="courseName"
                            placeholder="Course Name"
                            defaultValue={course.name}
                            onChange={handleChange}
                        />
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <FormInput
                            prepend="Code"
                            name="courseCode"
                            placeholder="(ex. DAT320)"
                            defaultValue={course.code}
                            onChange={handleChange}
                        />
                        <FormInput
                            prepend="Tag"
                            name="courseTag"
                            placeholder="(ex. Fall / Spring)"
                            defaultValue={course.tag}
                            onChange={handleChange}
                        />
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <FormInput
                            prepend="Slip days"
                            name="slipDays"
                            placeholder="(ex. 7)"
                            defaultValue={course.slipDays.toString()}
                            onChange={handleChange}
                            type="number"
                        />
                        <FormInput
                            prepend="Year"
                            name="courseYear"
                            placeholder="(ex. 2021)"
                            defaultValue={course.year.toString()}
                            onChange={handleChange}
                            type="number"
                        />
                    </div>
                    <div className="card-actions justify-end pt-4">
                        <button className="btn btn-primary" type="submit">
                            <i className="fa fa-save mr-2" />
                            Save Changes
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}

export default CourseForm
