import React from "react";
import { getFormattedDeadline } from "../Helpers";
import { useOvermind, useReaction } from "../overmind";
import { Submission } from "../proto/ag_pb";

const Status = {
    0: "NONE",
    1: "APPROVED",
    2: "REJECTED",
    3: "REVISION",
}

interface course {
    courseID: number
}

const LandingPageLabTable = (crs: course) => {
    const { state } = useOvermind()
    console.log(crs.courseID, "lol")
    //replace {} with a type of dictionary/record
    const makeTable = (): JSX.Element[] => {
        let table: JSX.Element[] = []
        let submission: Submission | undefined = undefined
        if (crs.courseID == 0) {
        for (const courseID in state.assignments) {
            let course = state.courses.find(course => course.getId() == Number(crs.courseID))
            state.assignments[courseID].forEach(assignment => {
                
                if(state.submissions[courseID]) {
                    submission = state.submissions[courseID][assignment.getOrder() - 1]
                    
                    
                
                if(submission){
                    table.push(
                        <tr key = {assignment.getId()} className= {"clickable-row "}>
                            <td>{course?.getCode()}</td>
                            <td>{assignment.getName()}</td>
                            <td>{submission.getScore()} / 100</td>
                            <td>{getFormattedDeadline(assignment.getDeadline())}</td>
                            <td>time left</td>
                            <td className={Status[submission.getStatus()]}>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}</td>
                            <td>{assignment.getIsgrouplab() ? "Yes": "No"}</td>
                        </tr>
                    )}
                }
            })
        }
        }
        else {
            let course = state.courses.find(course => course.getId() == Number(crs.courseID))
            state.assignments[crs.courseID].forEach(assignment => {
                
                if(state.submissions[crs.courseID]) {
                    submission = state.submissions[crs.courseID][assignment.getOrder() - 1]
                    
                    
                
                if(submission){
                    table.push(
                        <tr key = {assignment.getId()} className= {`clickable-row ${state.theme}`}>
                            <td>{assignment.getName()}</td>
                            <td>{submission.getScore()} / 100</td>
                            <td>{getFormattedDeadline(assignment.getDeadline())}</td>
                            <td>time left</td>
                            <td className={Status[submission.getStatus()]}>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}</td>
                            <td>{assignment.getIsgrouplab() ? "Yes": "No"}</td>
                        </tr>
                    )}
                }
            })
        }
        return table

    }
    
    return (
        <div>
            <table className="table table-curved" id="LandingPageTable">
                <thead>
                    <tr>
                        {crs.courseID == 0 ? <th>Course</th> : ""}
                        <th>Assignment</th>
                        <th>Progress</th>
                        <th>Deadline</th>
                        <th>Due in</th>
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