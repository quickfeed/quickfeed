import React, { useState } from "react"
import { useActions } from "../../overmind"
import { Course } from "../../../proto/ag/ag_pb"
import { json } from "overmind"




export const CourseForm = ({editCourse}: {editCourse?: Course}) => {
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
        <div className="box">
            <div hidden={editCourse ? true : false}>Organization Name
            <div><input onKeyUp={e => setOrgName(e.currentTarget.value)}></input><div className="btn btn-primary" onClick={() => actions.getOrganization(orgName)}>Check</div></div>
            </div>
            <form className="form-group well" style={{width: "400px"}} onSubmit={e => {e.preventDefault(), submitHandler()}}>
                <label htmlFor={"name"}>Course Name</label>
                <input  className="form-control" 
                        name="courseName" 
                        type="text" 
                        placeholder={"Course Name"} 
                        defaultValue={editCourse?.getName()} 
                        onChange={e => handleChange(e)}/>

                <label htmlFor={"email"}>Course Code</label>
                <input  className="form-control" 
                        name="courseCode" 
                        type="text" 
                        placeholder={"Tag (ex. DAT320)"} 
                        defaultValue={editCourse?.getCode()} 
                        onChange={e => handleChange(e)} />
               
                <label htmlFor={"email"}>Course Tag</label>
                <input  className="form-control" 
                        name="courseTag" 
                        type="text" 
                        placeholder={"Tag (ex. Fall / Spring)"} 
                        defaultValue={editCourse?.getTag()} 
                        onChange={e => handleChange(e)} />
                
                <label htmlFor={"studentid"}>Slip Days</label>
                <input  className="form-control" 
                        name="slipDays" 
                        type="number" 
                        placeholder={"Days (ex. 7)"} 
                        defaultValue={editCourse?.getSlipdays()} 
                        onChange={e => handleChange(e)}  />
                        
                <label htmlFor={"studentid"}>Year</label>
                <input className="form-control" name="courseYear" type="number" placeholder={"Year (ex. 2021)"} defaultValue={editCourse?.getYear()} onChange={e => handleChange(e)}  />
                <input className="btn btn-primary" type="submit" value={editCourse ? "Edit Course" : "Create Course"} style={{marginTop:"20px"}}/>
            </form>
        </div>
        
    )
}

export default CourseForm