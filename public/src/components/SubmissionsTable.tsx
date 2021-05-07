import React, { useEffect } from "react";
import { useHistory } from "react-router";
import { getFormattedDeadline, layoutTime, timeFormatter } from "../Helpers";
import { useOvermind, useReaction } from "../overmind";
import { Assignment, Course, Submission } from "../../proto/ag_pb";

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
/* Giving a courseID of zero (0) makes it display ALL labs for all courses, whereas providing a courseID displays labs for ONLY ONE course */
const SubmissionsTable = (crs: course) => {
    const { state } = useOvermind()
    const history  = useHistory()
    
    const redirectToLab = (courseid:number,assignmentid:number) => {
        history.push(`/course/${courseid}/${assignmentid}`)
    }


    const row = (assignment: Assignment, submission: Submission, course?: Course): JSX.Element => {
        const timeofDeadline = new Date(assignment.getDeadline())
        let time2Deadline = timeFormatter(timeofDeadline.getTime(),state.timeNow)
        return (
            <tr key={assignment.getId()} className={"clickable-row " + time2Deadline[1]} onClick={()=>redirectToLab(assignment.getCourseid(),assignment.getId())}>
            
            {crs.courseID==0 &&
                <td>{course?.getCode()}</td>
            }
                <td>{assignment.getName()}</td>
                <td>{submission.getScore()} / 100</td>
                <td>{getFormattedDeadline(assignment.getDeadline())}</td>
                <td>{time2Deadline[0] ? time2Deadline[2]: '--'}</td>
                <td className={Status[submission.getStatus()]}>{(assignment.getAutoapprove()==false && submission.getScore()>= assignment.getScorelimit()) ? "Awating approval":(assignment.getAutoapprove()==true && submission.getScore()>= assignment.getScorelimit())? "Approved(Auto approve)(shouldn't be in final version)":"Score not high enough"}</td>
                <td>{assignment.getIsgrouplab() ? "Yes": "No"}</td>
            </tr>
        )
    }

    const MakeLabTable = (): JSX.Element[] => {
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
                    {MakeLabTable()}
                </tbody>
            </table>
        </div>
    )
}

export default SubmissionsTable