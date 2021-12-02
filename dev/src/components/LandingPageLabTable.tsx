import React from "react";
import { useHistory } from "react-router";
import { generateStatusText, getFormattedTime, SubmissionStatus, timeFormatter } from "../Helpers";
import { useAppState } from "../overmind";
import { Assignment, Submission } from "../../proto/ag/ag_pb";


interface course {
    courseID: number
}

//** This component takes a courseID (number) to render a table containing lab information
/* Giving a courseID of zero (0) makes it display ALL labs for all courses, whereas providing a courseID displays labs for ONLY ONE course */

// TODO: Refactor this
const LandingPageLabTable = (crs: course): JSX.Element => {
    const state = useAppState()
    const history  = useHistory()
    
    const redirectToLab = (courseid: number, assignmentid: number) => {
        history.push(`/course/${courseid}/${assignmentid}`)
    }

    const MakeLabTable = (): JSX.Element[] => {
        const table: JSX.Element[] = []
        let submission: Submission = new Submission()
        for (const id in state.assignments) {
            // Use the index provided by the for loop if courseID provided == 0, else select the given course
            const courseID = crs.courseID > 0 ? crs.courseID : Number(id)
            const course = state.courses.find(course => course.getId() == courseID)  
            state.assignments[courseID]?.forEach(assignment => {
                if(state.submissions[courseID]) {
                    // Submissions are indexed by the assignment order - 1.
                    submission = state.submissions[courseID][assignment.getOrder() - 1]
                    if (!submission){
                        submission = new Submission()
                    }
                    if (submission.getStatus() > Submission.Status.APPROVED || submission.getStatus() < Submission.Status.APPROVED){
                        const deadline = timeFormatter(assignment.getDeadline(), state.timeNow)
                        if(deadline.daysUntil > 3 && submission.getScore() >= assignment.getScorelimit()) {
                            deadline.className = "table-success"
                        }
                        if(deadline.message){
                            table.push(
                                <tr key={assignment.getId()} className={"clickable-row " + deadline.className} onClick={()=>redirectToLab(courseID, assignment.getId())}>
                                    {crs.courseID == 0 && <th scope="row">{course?.getCode()}</th>}
                                    <td>{assignment.getName()}</td>
                                    <td>{submission.getScore()} / 100</td>
                                    <td>{getFormattedTime(assignment.getDeadline())}</td>
                                    <td>{deadline.message ? deadline.message : '--'}</td>
                                    <td className={SubmissionStatus[submission.getStatus()]}>
                                        {generateStatusText(assignment, submission)}
                                    </td>
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