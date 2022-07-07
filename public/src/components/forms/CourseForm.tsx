import React, { useState } from "react"
import { useActions } from "../../overmind"
import { Course, Organization } from "../../../proto/qf/qf_pb"
import FormInput from "./FormInput"
import CourseCreationInfo from "../admin/CourseCreationInfo"
import { useHistory } from "react-router"
import { defaultTag, defaultYear } from "../../Helpers"
import { Converter } from "../../convert"


// TODO: There are currently issues with navigating a new course without refreshing the page to trigger a state reload.
// TODO: Plenty of required fields are undefined across multiple components.

/** CourseForm is used to create a new course or edit an existing course.
 *  If `editCourse` is provided, the existing course will be modified.
 *  If no course is provided, a new course will be created. */
const CourseForm = ({ editCourse }: { editCourse?: Course.AsObject }): JSX.Element | null => {
    const actions = useActions()
    const history = useHistory()

    // TODO: This could go in go in a course-specific Overmind namespace rather than local state.
    // Local state for organization name to be checked against the server
    const [orgName, setOrgName] = useState("")
    const [org, setOrg] = useState<Organization>()

    // Local state containing the course to be created or edited (if any)
    const [course, setCourse] = useState(editCourse ? Converter.clone(editCourse) : Converter.create(Course))

    // Local state containing a boolean indicating whether the organization is valid. Courses that are being edited do not need to be validated.
    const [orgFound, setOrgFound] = useState<boolean>(editCourse ? true : false)

    /* Date object used to fill in certain default values for new courses */
    const date = new Date(Date.now())
    if (!editCourse) {
        course.year = defaultYear(date)
        course.tag = defaultTag(date)
    }

    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
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
                course.slipdays = Number(value)
                break
        }
        setCourse(course)
    }

    // Creates a new course if no course is being edited, otherwise updates the existing course
    const submitHandler = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        if (editCourse) {
            actions.editCourse({ course: course })
        } else {
            if (org) {
                const success = await actions.createCourse({ course: course, org: org })
                // If course creation was successful, redirect to the course page
                if (success) {
                    history.push("/courses")
                }
            } else {
                //org not found
            }
        }
    }

    // Trigger grpc call to check if org exists
    const getOrganization = async () => {
        const org = (await actions.getOrganization(orgName)).data
        if (org) {
            setOrg(org)
            setOrgFound(true)
        } else {
            setOrgFound(false)
        }
    }

    return (
        <div className="container">
            {editCourse ? null : <CourseCreationInfo />}
            <div className="row" hidden={editCourse ? true : false}>
                <div className="col input-group mb-3">
                    <div className="input-group-prepend">
                        <div className="input-group-text">Organization</div>
                    </div>
                    <input className="form-control" disabled={orgFound ? true : false} onKeyUp={e => setOrgName(e.currentTarget.value)} />
                    <span className={orgFound ? "btn btn-success disabled" : "btn btn-primary"} onClick={!orgFound ? () => getOrganization() : () => { return }}>
                        {orgFound ? <i className="fa fa-check" />: "Find"}
                    </span>
                </div>
            </div>
            {orgFound &&
                <form className="form-group" onSubmit={async e => await submitHandler(e)}>
                    <div className="row">
                        <FormInput prepend="Name"
                            name="courseName"
                            placeholder={"Course Name"}
                            defaultValue={editCourse?.name}
                            onChange={handleChange}
                        />
                    </div>
                    <div className="row">
                        <FormInput
                            prepend="Code"
                            name="courseCode"
                            placeholder={"(ex. DAT320)"}
                            defaultValue={editCourse?.code}
                            onChange={handleChange}
                        />
                        <FormInput
                            prepend="Tag"
                            name="courseTag"
                            placeholder={"(ex. Fall / Spring)"}
                            defaultValue={editCourse ? editCourse.tag : defaultTag(date)}
                            onChange={handleChange}
                        />
                    </div>
                    <div className="row">
                        <FormInput
                            prepend="Slip days"
                            name="slipDays"
                            placeholder={"(ex. 7)"}
                            defaultValue={editCourse?.slipdays.toString()}
                            onChange={handleChange}
                            type="number"
                        />
                        <FormInput
                            prepend="Year"
                            name="courseYear"
                            placeholder={"(ex. 2021)"}
                            defaultValue={editCourse ? editCourse.year.toString() : defaultYear(date).toString()}
                            onChange={handleChange}
                            type="number"
                        />
                    </div>
                    <input className="btn btn-primary" type="submit" value={editCourse ? "Edit Course" : "Create Course"} />
                </form>
            }
        </div>
    )
}

export default CourseForm
