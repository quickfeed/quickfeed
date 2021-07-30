import React from "react";
import { useHistory } from "react-router";
import { getFormattedTime, timeFormatter } from "../Helpers";
import { useAppState } from "../overmind";
import { Assignment, Course, Submission } from "../../proto/ag/ag_pb";

const Status = {
    0: "NONE",
    1: "APPROVED",
    2: "REJECTED",
    3: "REVISION",
}

interface course {
    courseID: number
    group?: boolean
}

//** This component takes a courseID (number) to render a table containing lab information
/*  Giving a courseID of zero (0) makes it display ALL labs for all courses, whereas providing a courseID displays labs for ONLY ONE course 
    Passing in group = true lists only group assignments
*/
const SubmissionsTable = (crs: course) => {
    const state = useAppState()
    const history  = useHistory()
    
    const redirectToLab = (courseid:number,assignmentid:number) => {
        history.push(`/course/${courseid}/${assignmentid}`)
    }


    const row = (assignment: Assignment, submission: Submission, course?: Course): JSX.Element => {
        let time2Deadline = timeFormatter(assignment.getDeadline(),state.timeNow)
        return (
            <tr key={assignment.getId()} className={"clickable-row " + time2Deadline[1]} onClick={()=>redirectToLab(assignment.getCourseid(),assignment.getId())}>
            
            {crs.courseID==0 &&
                <td>{course?.getCode()}</td>
            }
                <td>{assignment.getName()}</td>
                <td>{submission.getScore()} / 100</td>
                <td>{getFormattedTime(assignment.getDeadline())}</td>
                <td>{time2Deadline[0] ? time2Deadline[2]: '--'}</td>
                <td className={Status[submission.getStatus()]}>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}</td>
                <td>{assignment.getIsgrouplab() ? "Yes": "No"}</td>
            </tr>
        )
    }

    const LabTable: Function = (): JSX.Element[] => {
        let table: JSX.Element[] = []
        let submission: Submission | undefined = new Submission()
        submission.setScore(0)
            
        for (const courseID in state.assignments) {
            // Use the index provided by the for loop if courseID provided == 0, else select the given course
            let key = crs.courseID > 0 ? crs.courseID : Number(courseID)
            let course = state.courses.find(course => course.getId() == key)  
            state.assignments[key]?.forEach(assignment => {
                if(state.submissions[key]) {
                    // Submissions are indexed by the assignment order.
                    submission = state.submissions[key][assignment.getOrder() - 1]
                    if (submission === undefined ) {
                        submission = new Submission()
                    }
                    if (crs.group && submission.getGroupid() > 0) {
                        //Rewrite this to hide, this who are approved. if submission.getStatus() = 1 -> hide it.
                        table.push(
                            row(assignment, submission, course)        
                        )
                    }
                    if (!crs.group) {
                        table.push(
                            row(assignment, submission, course)        
                        )
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
            <table className="table table-curved" id="LandingPageTable">
                <thead>
                    <tr>
                        {crs.courseID !== 0 ? "" : <th>Course</th>}
                        <th>Assignment</th>
                        <th>Progress</th>
                        <th>Deadline</th>
                        <th>Due in</th>
                        <th>Status</th>
                        <th>Grouplab</th>
                    </tr>
                </thead>
                <tbody>
                    <LabTable />
                </tbody>
            </table>
        </div>
    )
}

export default SubmissionsTable