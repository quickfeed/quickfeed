import { action } from "overmind";
import React, {useCallback, useState, useEffect} from "react";
import { useOvermind, useReaction } from "../overmind";
import { Dict } from "../overmind/state";
import { Assignment } from "../proto/ag_pb";
import Home from "./Home";
import NavBar from './NavBar';




const LandingPageLabTable = () => {
    //TODO make this to inherit state/actions from Homepage.
    const { state , actions} = useOvermind()
    //replace {} with a type of dictionary/record
    const [assignments,setAssignments] = useState({})
    useEffect(() => {
        setAssignments(state.assignments)
        console.log(state.assignments)
    },[state.assignments])
    /*
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
    */
    const deadlineTable = Object.values(state.assignments).map(test =>{
        return(<h4 key={test.length}>{test}</h4>)
    })
    return (
        <div>
            
            {Object.values(assignments).map(arr =>{
                arr.map(assignment =>{
                    <h2 key={assignment.getId()}>{assignment.getName()} {console.log(assignment)}</h2>
                })
            })}
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
                    {deadlineTable}
                    
                </tbody>
            </table>
            
        </div>
    )
}

export default LandingPageLabTable