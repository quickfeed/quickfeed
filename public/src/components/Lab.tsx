import React, { useEffect } from 'react'
import { useParams } from 'react-router'
import { useOvermind } from '../overmind'
import LabResultTable from './LabResultTable'

interface MatchProps {
    lab: string
}
interface CourseID {
    crsID: number
}

const Lab = (props:CourseID) => {
    const { state } = useOvermind()
    const {lab} = useParams<MatchProps>()

    const getSubmission = state.submissions[props.crsID]?.map(submission => {
        if (submission.getAssignmentid() == Number(lab)) {
            const buildInfo = JSON.parse(submission.getBuildinfo())
            const prettyBuildlog = buildInfo.buildlog.split('\n').map((x: string, i: number) => <span key={i} >{x}<br /></span>)
            console.log(submission.getScoreobjects())
            return (
                <div key={submission.getId()}>
                    <LabResultTable id={submission.getAssignmentid()} courseID={props.crsID} />
                    <div className="well"><code>{prettyBuildlog}</code></div>
                    
                </div>
            )
        }
    })

    return (
        <div>
        {getSubmission}
        </div>
    )
}

export default Lab