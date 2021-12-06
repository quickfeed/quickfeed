import React, { useState } from "react"
import { useActions } from "../../overmind"
import { Course } from "../../../proto/ag/ag_pb"
import { json } from "overmind"
import FormInput from "./FormInput"




export const CourseForm = ({editCourse}: {editCourse?: Course}): JSX.Element => {
    const actions = useActions()
    const [orgName, setOrgName] = useState("")
    const [course, setCourse] = useState(editCourse ? json(editCourse) : new Course)


    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const {name, value} = event.currentTarget
        switch (name) {
            case "courseName":
                course.setName(value)
                break
            case "courseTag":
                course.setTag(value)
                break
            case "courseCode":
                course.setCode(value)
                break
            case "courseYear":
                course.setYear(Number(value))
                break
            case "slipDays":
                course.setSlipdays(Number(value))
                break
        }
        setCourse(course)
    }

    const submitHandler = () => {
        if (editCourse) {
            actions.editCourse({ course: course })
        } else {
            actions.createCourse({ course: course, orgName: orgName })
        }
    }

    return (
        <div className="container">
            <div className="jumbotron">
                <h1 className="display-4">Create Course</h1>
                <p className="lead">
                For each new semester of a course, QuickFeed requires a new GitHub organization. This is to keep the student roster for the different runs of the course separate.
                </p>
                <p><a className="badge-pill badge-success" href="https://github.com/account/organizations/new">Create an organization</a>  for your course. The course organization must allow private repositories.</p>
                <p>
                QuickFeed will create the following repositories for you:
                </p>
                <ul>
                    <li>info</li>
                    <li>assignments</li>
                    <li>tests</li>
                </ul>
                <p>Please refer to the <a className="badge-pill badge-primary" href="https://github.com/autograde/quickfeed/blob/master/doc/teacher.md">documentation</a> for further instructions on how to work with the various repositories.</p>
            </div>
            <div className="row" hidden={editCourse ? true : false}>
                <div className="col input-group mb-3">
                    <div className="input-group-prepend">
                        <div className="input-group-text">Organization</div>
                    </div>
                    <input className="form-control" onKeyUp={e => setOrgName(e.currentTarget.value)}></input>
                    <span className="btn btn-primary" onClick={() => actions.getOrganization(orgName)}>Find</span>
                </div>
            </div>
            <form className="form-group" onSubmit={e => {e.preventDefault(), submitHandler()}}>
                <div className="row">
                    <FormInput  prepend="Name"
                            name="courseName" 
                            placeholder={"Course Name"} 
                            defaultValue={editCourse?.getName()}
                            onChange={handleChange} 
                    />
                </div>
                <div className="row">
                    <FormInput
                            prepend="Code"
                            name="courseCode" 
                            placeholder={"(ex. DAT320)"} 
                            defaultValue={editCourse?.getCode()} 
                            onChange={handleChange} 
                    />
                    <FormInput
                            prepend="Tag"
                            name="courseTag" 
                            placeholder={"(ex. Fall / Spring)"} 
                            defaultValue={editCourse?.getTag()} 
                            onChange={handleChange} 
                    />
                </div>
                <div className="row">
                    <FormInput
                            prepend="Slip days"
                            name="slipDays" 
                            placeholder={"(ex. 7)"} 
                            defaultValue={editCourse?.getSlipdays().toString()} 
                            onChange={handleChange} 
                    />
                    <FormInput prepend="Year" name="courseYear" placeholder={"(ex. 2021)"} defaultValue={editCourse?.getYear().toString()}/>
                </div>
                <input className="btn btn-primary" type="submit" value={editCourse ? "Edit Course" : "Create Course"} style={{marginTop:"20px"}}/>
            </form>
        </div>
        
    )
}

export default CourseForm