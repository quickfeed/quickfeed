import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { useOvermind } from '../overmind'
import LabResultTable from './LabResultTable'
import ReviewResult from './ReviewResult'

interface MatchProps {
    id: string
    lab: string
}


const Lab = () => {
    const { state, actions } = useOvermind()
    const {id ,lab} = useParams<MatchProps>()
    const courseID = Number(id)
    const assignmentID = Number(lab)
    useEffect(() => {
        actions.setActiveLab(assignmentID)
        
        // Needs to handle what to do in case of no commit hash, such as manually graded submissions
        const t = setInterval(() => {  
            actions.getHash({courseID: courseID, assignmentID: assignmentID})
        }, 10000)
        return () => {clearInterval(t), actions.setActiveLab(-1)}
    }, [lab])

    const getSubmission = state.submissions[courseID]?.map(submission => {
        if (submission.getAssignmentid() == assignmentID) {
            const buildInfo = JSON.parse(submission.getBuildinfo())
            const prettyBuildlog = buildInfo.buildlog.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>);

            return (
                <div key={submission.getId()}>
                    <LabResultTable id={submission.getAssignmentid()} courseID={courseID} />
                    {state.assignments[courseID].find(a => a.getId() === assignmentID)?.getSkiptests() ? <ReviewResult review={submission.getReviewsList()}/> : ""}
                    <div className="card bg-light"><code className="card-body" style={{color: "#c7254e"}}>{prettyBuildlog}</code></div>
                </div>
            )
    }})

    return (
        <div className="col-md-8 box">
        {getSubmission}
        </div>
    )
}

export default Lab