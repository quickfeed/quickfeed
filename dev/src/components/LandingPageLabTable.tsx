import React from "react";
import { useHistory } from "react-router";
import { generateStatusText, getFormattedTime, SubmissionStatus, timeFormatter } from "../Helpers";
import { useAppState } from "../overmind";
import { Assignment, Submission } from "../../proto/ag/ag_pb";
import { Progress, ProgressBar } from "./ProgressBar";


//** This component takes a courseID (number) to render a table containing lab information
/* Giving a courseID of zero (0) makes it display ALL labs for all courses, whereas providing a courseID displays labs for ONLY ONE course */
// TODO: Refactor this
const LandingPageLabTable = (): JSX.Element => {
    const state = useAppState()
    const history  = useHistory()
    
    const redirectToLab = (courseid: number, assignmentid: number) => {
        history.push(`/course/${courseid}/${assignmentid}`)
    }

    const sortedAssignments = () => {
        const assignments: Assignment[] = []
        for (const courseID in state.assignments) {
            state.assignments[courseID].forEach(assignment => {
                assignments.push(assignment)
            })
        }
        assignments.sort((a, b) => {
            if (b.getDeadline() > a.getDeadline()) {
                return -1
            }
            if (a.getDeadline() > b.getDeadline()) {
                return 1
            }
            return 0
        })
        return assignments
    }

    const SubmissionsTable = (): JSX.Element[] => {
        const table: JSX.Element[] = []
        let submission: Submission = new Submission()
        sortedAssignments().forEach(assignment => {
            if (!state.submissions[assignment.getCourseid()]) {
                return
            }
            // Submissions are indexed by the assignment order - 1.
            submission = state.submissions[assignment.getCourseid()][assignment.getOrder() - 1]
            if (!submission){
                submission = new Submission()
            }
            if (submission.getStatus() > Submission.Status.APPROVED || submission.getStatus() < Submission.Status.APPROVED){
                const deadline = timeFormatter(assignment.getDeadline(), state.timeNow)
                if(deadline.daysUntil > 3 && submission.getScore() >= assignment.getScorelimit()) {
                    deadline.className = "table-success"
                }
                if(!deadline.message){
                    return
                }
                const course = state.courses.find(course => course.getId() === assignment.getCourseid())
                table.push(
                    <tr key={assignment.getId()} className={"clickable-row " + deadline.className} onClick={()=>redirectToLab(Number(assignment.getCourseid()), assignment.getId())}>
                        <th scope="row">{course?.getCode()}</th>
                        <td>{assignment.getName()}</td>
                        <td><ProgressBar assignmentIndex={assignment.getOrder() - 1} courseID={assignment.getCourseid()} submission={submission} type={Progress.OVERVIEW} /></td>
                        <td>{getFormattedTime(assignment.getDeadline())}</td>
                        <td>{deadline.message ? deadline.message : '--'}</td>
                        <td className={SubmissionStatus[submission.getStatus()]}>
                            {generateStatusText(assignment, submission)}
                        </td>
                        <td>{assignment.getIsgrouplab() ? "Yes": "No"}</td>
                    </tr>
                )
            }
        })
        return table   
    }

    return (
        <div>
            <table className="table rounded-lg table-bordered table-hover" id="LandingPageTable">
                <thead >
                    <tr>
                        <th scope="col">Course</th>
                        <th scope="col">Assignment</th>
                        <th scope="col">Progress</th>
                        <th scope="col">Deadline</th>
                        <th scope="col">Due in</th>
                        <th scope="col">Status</th>
                        <th scope="col">Grouplab</th>
                    </tr>
                </thead>
                <tbody>
                    {SubmissionsTable()}
                </tbody>
            </table>
        </div>
    )
}

export default LandingPageLabTable