import React, { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { useOvermind } from "../overmind"


export const ProgressBar = (props: {courseID: number, assignmentID: number, type: string}) => {
    const { state } = useOvermind()

    let percentage = 0
    let score = 0
    if (state.submissions[props.courseID][props.assignmentID] !== undefined) {
        let rand = Math.random()
        let submission = state.submissions[props.courseID][props.assignmentID]
        percentage = 100 - (submission.getScore() - rand * 100)
        score = submission.getScore() - rand * 100
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
    return (
        <div>
            
        </div>
    )
}
