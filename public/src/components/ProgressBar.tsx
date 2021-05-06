import React from "react"
import { useOvermind } from "../overmind"
import { Assignment, Submission } from "../../proto/ag_pb"


export const ProgressBar = (props: {courseID: number, assignmentIndex: number, submission?: Submission, type: string}) => {
    const { state } = useOvermind()
    let submission:Submission = props?.submission !== undefined ? props?.submission : new Submission()
    let percentage = 0
    let score = 0
    if (state.submissions[props.courseID] !== undefined) {
        let rand = Math.random()
        let submission = state.submissions[props.courseID][props.assignmentIndex]
        percentage = 100 - (submission.getScore() - rand * 100)
        score = submission.getScore() - rand * 100
    }

    if(props.type === "navbar") {
        return (
            <div style={{ 
                position: "absolute", 
                borderBottom: "2px solid green", 
                bottom: 0, 
                left: 0, 
                right: `${percentage}%`, 
                borderColor: `${score >= state.assignments[props.courseID][props.assignmentIndex].getScorelimit() ? "green" : "yellow"}`
                , opacity: 0.3 }}>
    
                </div>
        )
    }
    if(props.type === "lab") {
        let color = "bg-success"
        if (submission.getStatus()==0){
            if(submission.getScore()>=state.assignments[props.courseID][props.assignmentIndex].getScorelimit()){
                color = "bg-primary"
            }else{
                color = "bg-secondary"
            }

        }
        //Not completed
        if (submission.getStatus()==2){
            color = "bg-danger"
        }
        if (submission.getStatus()==3){
            color = "bg-warning text-dark"
        }
        return (
            <div className="progress">
                <div className={"progress-bar "+color} role="progressbar" style={{width: props.submission?.getScore() + "%", transitionDelay: "0.5s"}} aria-valuenow={submission.getScore()} aria-valuemin={0} aria-valuemax={100}>{props.submission?.getScore()}%</div>
            </div>
        )
    }
    return (
        <div>
            
        </div>
    )
}
