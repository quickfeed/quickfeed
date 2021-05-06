import React, { useEffect, useState } from "react"
import { useOvermind } from "../../overmind"
import { Course, Enrollment } from "../../proto/ag_pb"
import CourseCard from "../CourseCard"




export const CourseCreationForm = () => {
    const {actions} = useOvermind()
    const [orgName, setOrgName] = useState("")
    const [course, setCourse] = useState(new Course)


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
        actions.createCourse({course: course, orgName: orgName})
    }

    return (
        <div className="box">
            <div>Organization Name</div>
            <div><input onKeyUp={e => setOrgName(e.currentTarget.value)}></input><div className="btn btn-primary" onClick={() => actions.getOrganization(orgName)}>Check</div></div>
            <form className="form-group well" style={{width: "400px"}} onSubmit={e => {e.preventDefault(), submitHandler()}}>
                        <label htmlFor={"name"}>Course Name</label>
                        <input className="form-control" name="courseName" type="text" placeholder={"Course Name"} onChange={e => handleChange(e)}/>
                        <label htmlFor={"email"}>Course Code</label>
                        <input className="form-control" name="courseCode" type="text" placeholder={"Tag (ex. DAT320)"} onChange={e => handleChange(e)} />
                        <label htmlFor={"email"}>Course Tag</label>
                        <input className="form-control" name="courseTag" type="text" placeholder={"Tag (ex. Fall / Spring)"} onChange={e => handleChange(e)} />
                        <label htmlFor={"studentid"}>Slip Days</label>
                        <input className="form-control" name="slipDays" type="number" placeholder={"Days (ex. 7)"} onChange={e => handleChange(e)}  />
                        <label htmlFor={"studentid"}>Year</label>
                        <input className="form-control" name="courseYear" type="number" placeholder={"Year (ex. 2021)"} onChange={e => handleChange(e)}  />
                        <input className="btn btn-primary" type="submit" value="Create Course" style={{marginTop:"20px"}}/>
            </form>
        </div>
        
    )
}

export default CourseCreationForm