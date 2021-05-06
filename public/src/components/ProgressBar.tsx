import React from "react"
import { useOvermind } from "../overmind"
import { Submission } from "../proto/ag_pb"


export const ProgressBar = (props: {courseID: number, assignmentID: number, submission?: Submission, type: string}) => {
    const { state } = useOvermind()

    let percentage = 0
    let score = 0
    if (state.submissions[props.courseID][props.assignmentID] !== undefined) {
        let submission = state.submissions[props.courseID][props.assignmentID]
        percentage = 100 - submission.getScore()
        score = submission.getScore()* 100
    }

    if(props.type === "navbar") {
        return (
            <div style={{ 
                position: "absolute", 
                borderBottom: "1px solid green", 
                bottom: 0, 
                left: 0, 
                right: `${percentage}%`, 
                borderColor: `${score >= state.assignments[props.courseID][props.assignmentID].getScorelimit() ? "green" : "yellow"}`
                , opacity: 0.3 }}>
    
                </div>
        )
    }
    if(props.type === "lab") {
        return (
            <div className="progress">
                <div className="progress-bar bg-success" role="progressbar" style={{width: props.submission?.getScore() + "%", transitionDelay: "0.5s"}} aria-valuenow={state.submissions[props.courseID][props.assignmentID]?.getScore()} aria-valuemin={0} aria-valuemax={100}>{props.submission?.getScore()}%</div>
            </div>
        )
    }
    return (
        <div>
            
        </div>
    )
}
