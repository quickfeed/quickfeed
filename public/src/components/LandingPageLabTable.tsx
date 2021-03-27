import { type } from "os";
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

//** This component takes a courseID (number) to render a table containing lab information
/* Giving a courseID of zero (0) makes it display ALL labs for all courses, whereas providing a courseID displays labs for ONLY ONE course */
const LandingPageLabTable = (crs: course) => {
    const { state } = useOvermind()
    

    const MakeLabTable = (): JSX.Element[] => {
        let table: JSX.Element[] = []
        let submission: Submission | undefined = undefined

            for (const courseID in state.assignments) {
                // Use the index provided by the for loop if courseID provided == 0, else select the given course
                let index = crs.courseID > 0 ? crs.courseID : Number(courseID)
                let course = state.courses.find(course => course.getId() == index)  

                state.assignments[index].forEach(assignment => {
                    
                    if(state.submissions[courseID]) {
                        // Submissions are indexed by the assignment order.
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

                // Break out of the for loop on the first iteration if we are only rendering information for one course
                if (crs.courseID > 0) {
                    break
                }
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
                    {MakeLabTable()}
                </tbody>
            </table>
        </div>
    )
}

export default LandingPageLabTable