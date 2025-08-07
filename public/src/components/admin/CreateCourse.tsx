import React, { useState } from "react"
import { Course } from "../../../proto/qf/types_pb"
import { useActions, useAppState } from "../../overmind"
import CourseForm from "../forms/CourseForm"
import CourseCreationInfo from "./CourseCreationInfo"
import { Color } from "../../Helpers"


const CreateCourse = () => {
    const state = useAppState()
    const actions = useActions().global
    const [course, setCourse] = useState<Course>()
    const [orgName, setOrgName] = useState("")

    const refresh = React.useCallback(async () => {
        await actions.getCourses()
        const c = state.courses.find(c => c.ScmOrganizationName === orgName)
        if (c) {
            await actions.getEnrollmentsByUser()
            setCourse(c)
        } else {
            actions.alert({ text: "Course not found. Make sure the organization name is correct and that you have installed the GitHub App.", color: Color.YELLOW, delay: 10000 })
        }
    }, [actions, orgName, state.courses])

    const buttonClass = course ? "btn btn-success disabled" : "btn btn-primary"
    const findTextOrIcon = course ? <i className="fa fa-check" /> : "Find"
    const refreshIfNoCourse = course ? undefined : refresh
    return (
        <div className="container">
            <CourseCreationInfo />
            <div className="row">
                <div className="col input-group mb-3">
                    <div className="input-group-prepend">
                        <div className="input-group-text">Get Course</div>
                    </div>
                    <input className="form-control" disabled={course ? true : false} onKeyUp={e => setOrgName(e.currentTarget.value)} />
                    <span role="button" aria-hidden="true" className={buttonClass} onClick={refreshIfNoCourse}>
                        {findTextOrIcon}
                    </span>
                </div>
            </div>
            {course ? <CourseForm courseToEdit={course} /> : null}
        </div>
    )
}

export default CreateCourse
