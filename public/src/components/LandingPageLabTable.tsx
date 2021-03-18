import { render } from "@testing-library/react";
import { useOvermind, useReaction } from "../overmind";
import { Assignment, Submission } from "../proto/ag_pb";



interface DictionaryProps{
    submissions : {
        [courseid:number] : Submission[]
    }
    assignments : {
        [courseid:number] : Assignment[]
    }
}

let LandingPageLabTable = (props:DictionaryProps) => {
    //TODO make this to inherit state/actions from Homepage.
    const { state , actions} = useOvermind()
    
    //replace {} with a type of dictionary/record
    console.log(props.assignments,props.submissions)
    const tableMap = Object.entries(props.assignments).map(([crsid,assignments]) => {
        const now = new Date()
        let courseid = Number(crsid)
        let crsName = state.courses.find(course => course.getId() === courseid)?.getCode()
        return assignments.map((assignment)=>{
            let submission = props.submissions[courseid].find(submission => submission.getAssignmentid() === assignment.getId())
            let deadline = new Date(assignment.getDeadline())
            if(submission && !submission?.getStatus()){
                return(
                    <tr key = {assignment.getId()} className= {"clickable-row "}>
                        <td>{crsName}</td>
                        <td>{assignment.getName()}</td>
                        <td>{submission.getScore()} / {assignment.getScorelimit()}</td>
                        <td>{deadline.getDate()}-{deadline.getMonth()}-{deadline.getFullYear()}|{deadline.getHours()}:{deadline.getMinutes()}</td>
                        <td></td>
                        <td>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}</td>
                        <td>{Boolean(assignment.getIsgrouplab()) ? "Yes": "No"}</td>
                    </tr>
                )    
            }
        })
        
    })
    
    return (
        <div>
            <table className="table" id="LandingPageTable">
                <thead>
                    <tr>
                        <th>Course</th>
                        <th>Assignment Name</th>
                        <th>Progress</th>
                        <th>Deadline</th>
                        <th>Time Left</th>
                        <th>Status</th>
                        <th>Grouplab</th>
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