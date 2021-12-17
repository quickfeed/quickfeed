import React, { useState } from "react"
import { useActions } from "../../overmind"
import { Course } from "../../../proto/ag/ag_pb"
import { json } from "overmind"
import FormInput from "./FormInput"
import CourseCreationInfo from "../admin/CourseCreationInfo"




export const CourseForm = ({editCourse}: {editCourse?: Course}): JSX.Element => {
    const actions = useActions()
    const [orgName, setOrgName] = useState("")
    const [course, setCourse] = useState(editCourse ? json(editCourse) : new Course)
    const [orgFound, setOrgFound] = useState<boolean>(false)
    /* Date object used to fill in certain default values for new courses */
    const date = new Date(Date.now())

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

    const getOrganization = async () => {
        setOrgFound(await actions.getOrganization(orgName))
    }

    return (
        <div className="container">
            {editCourse ? <></> : <CourseCreationInfo />}
            <div className="row" hidden={editCourse ? true : false}>
                <div className="col input-group mb-3">
                    <div className="input-group-prepend">
                        <div className="input-group-text">Organization</div>
                    </div>
                    <input className="form-control" onKeyUp={e => setOrgName(e.currentTarget.value)}></input>
                    <span className={orgFound ? "btn btn-success disabled" : "btn btn-primary"} onClick={!orgFound ? () => getOrganization() : () => {return}}>
                        {orgFound ? <i className="fas fa-check"></i> :  "Find"}
                    </span>
                </div>
            </div>
            {orgFound &&
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
                            defaultValue={editCourse ? editCourse.getTag() : defaultTag(date)} 
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
                    <FormInput prepend="Year" name="courseYear" placeholder={"(ex. 2021)"} defaultValue={editCourse ? editCourse.getYear().toString() : defaultYear(date)}/>
                </div>
                <input className="btn btn-primary" type="submit" value={editCourse ? "Edit Course" : "Create Course"}/>
            </form>
            }
        </div>
        
    )
}

export default CourseForm

const defaultTag = (date: Date) => {
    return date.getMonth() >= 10 && date.getMonth() < 4 ? "Spring" : "Fall"
}

const defaultYear = (date: Date) => {
    return (date.getMonth() <= 11 && date.getDate() <= 31) && date.getMonth() > 10 ? (date.getFullYear() + 1).toString() : date.getFullYear().toString()
}