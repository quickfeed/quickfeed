import React, { useEffect } from 'react'
import { RouteComponentProps, useParams, useRouteMatch } from 'react-router'
import { useActions, useOvermind } from '../overmind'
import { Submission } from "../proto/ag_pb"


interface MatchProps {
    lab: string
}

const Lab = () => {
    const { state } = useOvermind()
    const {lab} = useParams<MatchProps>()

    const getSubmission = state.submissions.map(submission => {
        if (submission.getAssignmentid() == Number(lab)) {
            console.log(submission.getId())
            return (
                <div key={submission.getId()}>
                    <h1>{submission.getScore()}%</h1>
                    <code>{submission.getBuildinfo()}</code>
                </div>
            )
        }
    })

    return (
        <div>
        Lab:
        {getSubmission}
        </div>
    )
}

export default Lab