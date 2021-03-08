import { action } from "overmind";
import React, {useCallback, useState, useEffect} from "react";
import { useOvermind, useReaction } from "../overmind";
import Home from "./Home";
import NavBar from './NavBar';



const LandingPageLabTable = () => {
    //TODO make this to inherit state/actions from Homepage.
    const { state , actions} = useOvermind()
    /*
    useEffect(() => {
       state.enrollments.map(enrol =>{
           actions.getSubmissions(enrol.getCourseid())
       })
        //console.log(state.enrollments)
    },[])
    */
    const tableMap = state.assignments.map(assignment => {
        const deadline = new Date(assignment.getDeadline())
        const now = new Date()
        const crscode= state.courses.find(course => course.getId() ===assignment.getCourseid())?.getCode()
        //This statement is just for testing, this current logic, wouldn't give you assignments where the deadline has passed.
        //Need to rethink this.
        const submission = state.submissions.find(submission => submission.getAssignmentid() === assignment.getId())
        if (submission){
            return(
                <tr key={assignment.getId()}>
                    <td>
                        {crscode}
                    </td>
                    <td>
                        {assignment.getName()}
                    </td>
                    <td>
                        {submission.getScore()} / {assignment.getScorelimit()}
                    </td>
                    <td>
                    {deadline.getDate()}-{deadline.getMonth()}-{deadline.getFullYear()}|{deadline.getHours()}:{deadline.getMinutes()}
                    </td>
                    <td>
                        timeleft
                    </td>
                    <td>
                        {(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}
                    </td>
                </tr>
                
            )
        }
    })
    

    return (
        <div>
            <table className="table">
                <thead>
                    <tr>
                        <th>Course</th>
                        <th>Assignment Name</th>
                        <th>Progress</th>
                        <th>Deadline</th>
                        <th>Time Left</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {tableMap}
                </tbody>
            </table>
            
        </div>
    )
}

export default LandingPageLabTable