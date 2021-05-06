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

    useEffect(() => {
        //actions.sHash({courseID: Number(id), assignmentID: Number(lab)})
        //const t = setInterval(() => {
        //    actions.getHash({courseID: Number(id), assignmentID: Number(lab)})
        //}, 10000)
        //return () => clearInterval(t)
        actions.setActiveLab(Number(lab))
        return () => actions.setActiveLab(-1)
    }, [lab])

    const getSubmission = state.submissions[Number(id)]?.map(submission => {
        if (submission.getAssignmentid() == Number(lab)) {
            
            const buildInfo = JSON.parse(submission.getBuildinfo())
            const prettyBuildlog = buildInfo.buildlog.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>);

            return (
                <div key={submission.getId()}>
                    <LabResultTable id={submission.getAssignmentid()} courseID={Number(id)} />
                    {state.assignments[Number(id)].find(a => a.getId() === submission.getAssignmentid())?.getSkiptests() ? <ReviewResult review={submission.getReviewsList()}/> : ""}
                    <div className="card bg-light"><code className="card-body" style={{color: "#c7254e"}}>{prettyBuildlog}</code></div>
                </div>
            )
    }})

    return (
        <div className="container box">
        {getSubmission}
        </div>
    )
}

export default Lab