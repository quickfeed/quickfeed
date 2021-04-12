import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { useOvermind } from '../overmind'
import LabResultTable from './LabResultTable'

interface MatchProps {
    id: string
    lab: string
}


const Lab = () => {
    const { state } = useOvermind()
    const {id ,lab} = useParams<MatchProps>()

    const getSubmission = state.submissions[Number(id)]?.map(submission => {
        if (submission.getAssignmentid() == Number(lab)) {
            const buildInfo = JSON.parse(submission.getBuildinfo())
            
            const prettyBuildlog = buildInfo.buildlog.split("\n").map((x: string, i: number) => <span key={i} >{x}<br /></span>);
            console.log(JSON.parse(buildInfo.buildlog))
            return (
                <div key={submission.getId()}>
                    <LabResultTable id={submission.getAssignmentid()} courseID={Number(id)} />
                    <div className="well"><code>{prettyBuildlog}</code></div>
                    
                </div>
            )
        }
    })

    return (
        <div className="box">
        {getSubmission}
        </div>
    )
}

export default Lab