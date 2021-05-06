import { useHistory } from "react-router"
import { getFormattedDeadline, SubmissionStatus } from "../Helpers"
import { useOvermind } from "../overmind"
import { Submission } from "../../proto/ag_pb"
import { ProgressBar } from "./ProgressBar"

interface MatchProps {
    crsid: number
}

export const CourseLabs = (props:MatchProps) =>  {
    const { state } = useOvermind()
    const history  = useHistory()
    
    function redirectToLab(assignmentid:number){
        history.push(`/course/${props.crsid}/${assignmentid}`)
    }
    const makeListItems =  ():JSX.Element[] =>{
        let listItems:JSX.Element[] = []
        let submission: Submission = new Submission()
        if(state.assignments[props.crsid]!==undefined){
            state.assignments[props.crsid].forEach(assignment => {
                if(state.submissions[props.crsid]!==undefined) {
                    // Submissions are indexed by the assignment order.
                    
                    if (state.submissions[props.crsid][assignment.getOrder() - 1]!==undefined){
                        submission = state.submissions[props.crsid][assignment.getOrder() - 1]
                    }
                }
                listItems.push(
                    <li key={assignment.getId()} className="list-group-item border"style={{marginBottom:"5px",cursor:"pointer"}} onClick={()=>redirectToLab(assignment.getId())}>
                        
                        <div className="row" >
                            <div className="col-8"><strong>{assignment.getName()}</strong></div>
                            <div className="col-4 text-center"><strong>Deadline:</strong></div>
                        </div>
                        <div className="row" >
                            <div className="col-5"><ProgressBar courseID={props.crsid} assignmentIndex={assignment.getOrder()-1} submission={submission} type="lab"/></div>
                            <div className="col-3 text-center">{(submission.getStatus()==0 && submission.getScore()>=assignment.getScorelimit()) ? "AWAITING APPROVAL":SubmissionStatus[submission.getStatus()]}</div>
                            <div className="col-4 text-center">{getFormattedDeadline(assignment.getDeadline())}</div>
                        </div>
                    </li>
                )
            })
        }
        return listItems
    }
    return (
        <ul className="list-group">
            {makeListItems()}
        </ul>
    )
}