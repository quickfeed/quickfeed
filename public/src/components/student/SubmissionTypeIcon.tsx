import React from 'react'

interface SubmissionTypeIconProps {
    solo: boolean
}

const SubmissionTypeIcon: React.FC<SubmissionTypeIconProps> = ({ solo }) => {
    const indicator = solo ? "fa-user" : "fa-users"
    return (
        <i
            className={`fa ${indicator} submission-icon`}
            title={`${solo ? "Solo" : "Group"} submission`}
        />
    )
}

export default SubmissionTypeIcon
