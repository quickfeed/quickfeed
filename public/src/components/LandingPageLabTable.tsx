import { render } from "@testing-library/react";
import { useEffect } from "react";
import { useOvermind, useReaction } from "../overmind";
import { Assignment, Submission } from "../proto/ag_pb";
import { getFormattedDeadline } from "../Helpers"


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
    
    const makeTable = (): JSX.Element[] => {
        let table: JSX.Element[] = []
        for (const courseID in state.assignments) {
            let crsName = state.courses.find(course => course.getId() === Number(courseID))?.getCode()
            state.assignments[courseID].map(assignment => {
                if(state.submissions[courseID]) {
                    let submission = state.submissions[courseID].find(submission => assignment.getId() === submission.getAssignmentid())
                
                if(submission){
                table.push(
                    <tr key = {assignment.getId()} className= {"clickable-row "}>
                        <td>{crsName}</td>
                        <td>{assignment.getName()}</td>
                        <td>{submission.getScore()} / {assignment.getScorelimit()}</td>
                        <td>{getFormattedDeadline(assignment.getDeadline())}</td>
                        <td></td>
                        <td>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}</td>
                        <td>{Boolean(assignment.getIsgrouplab()) ? "Yes": "No"}</td>
                    </tr>
                )
                }
            }
            })
        }
        return table
    }
    //replace {} with a type of dictionary/record
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
                    {makeTable()}
                </tbody>
            </table>
        </div>
    )
}

export default LandingPageLabTable