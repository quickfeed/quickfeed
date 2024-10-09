import React from 'react'
import { Assignment, Submission } from "../../../proto/qf/types_pb"

interface SubmissionTypeIconProps {
  assignment: Assignment
  submission: Submission
}

const SubmissionTypeIcon: React.FC<SubmissionTypeIconProps> = ({ assignment, submission }) => {
  const isGroupLab = assignment.isGroupLab
  const isSoloSubmission = submission.userID !== 0n
  const solo = (isGroupLab && isSoloSubmission) || !isGroupLab
  const indicator = solo ? "fa-user" : "fa-users"

  if (assignment.isGroupLab) {
    return (
      <i
        className={`fa ${indicator} submission-icon`}
        title={solo ? "Solo submission" : "Group submission"}
      />
    )
  }
  return null
}

export default SubmissionTypeIcon
