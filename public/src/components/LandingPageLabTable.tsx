import React, { useEffect } from "react";
import { useHistory } from "react-router";
import { getFormattedTime, layoutTime, SubmissionStatus, timeFormatter } from "../Helpers";
import { useOvermind, useReaction } from "../overmind";
import { Submission } from "../../proto/ag/ag_pb";


interface course {
    courseID: number
}

//** This component takes a courseID (number) to render a table containing lab information
/* Giving a courseID of zero (0) makes it display ALL labs for all courses, whereas providing a courseID displays labs for ONLY ONE course */
const LandingPageLabTable = (crs: course) => {
    const { state } = useOvermind()
    const history  = useHistory()
    
    function redirectToLab(courseid:number,assignmentid:number){
        history.push(`/course/${courseid}/${assignmentid}`)
    }

    const MakeLabTable = (): JSX.Element[] => {
        let table: JSX.Element[] = []
        let submission: Submission = new Submission()
            for (const courseID in state.assignments) {
                // Use the index provided by the for loop if courseID provided == 0, else select the given course
                let key = crs.courseID > 0 ? crs.courseID : Number(courseID)
                let course = state.courses.find(course => course.getId() == key)  
                state.assignments[key]?.forEach(assignment => {
                    if(state.submissions[key]) {
                        // Submissions are indexed by the assignment order.
                        submission = state.submissions[key][assignment.getOrder() - 1]
                        if (submission===undefined){
                            submission = new Submission()
                        }
                        if (submission.getStatus() !== Submission.Status.APPROVED){
                            let time2Deadline = timeFormatter(assignment.getDeadline(),state.timeNow)
                            if(time2Deadline[3] >3 && (submission.getScore() >= assignment.getScorelimit() &&(submission.getStatus()<1))){
                                time2Deadline[1]= "table-success"
                            }
                            if(time2Deadline[0]){
                                table.push(
                                    <tr key={assignment.getId()} className={"clickable-row " + time2Deadline[1]} onClick={()=>redirectToLab(assignment.getCourseid(),assignment.getId())}>
                                        {crs.courseID==0 &&
                                        <th scope="row">{course?.getCode()}</th>
                                        }
                                        <td>{assignment.getName()}</td>
                                        <td>{submission.getScore()} / 100</td>
                                        <td>{getFormattedTime(assignment.getDeadline())}</td>
                                        <td>{time2Deadline[0] ? time2Deadline[2]: '--'}</td>
                                        <td className={SubmissionStatus[submission.getStatus()]}>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit() && submission.getStatus()<1)? "Awaiting approval":(submission.getScore()>=assignment.getScorelimit()? SubmissionStatus[submission.getStatus()] :"Score not high enough")}</td>
                                        <td>{assignment.getIsgrouplab() ? "Yes": "No"}</td>
                                    </tr>
                                )
                            }
                        }
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
            <table className="table rounded-lg table-bordered table-hover" id="LandingPageTable">
                <thead >
                    <tr>
                        {crs.courseID !== 0 ? null : <th scope="col">Course</th>}
                        <th scope="col">Assignment</th>
                        <th scope="col">Progress</th>
                        <th scope="col">Deadline</th>
                        <th scope="col">Due in</th>
                        <th scope="col">Status</th>
                        <th scope="col">Grouplab</th>
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