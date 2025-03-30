import React, { useCallback } from 'react'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { getFormattedTime } from "../../Helpers"
import SubmissionRow from './SubmissionRow'
import { useHistory } from 'react-router'

interface AssignmentCardProps {
  assignment: Assignment
  submissions: Submission[]
  courseID: string
  selfID: bigint
}

const AssignmentCard: React.FC<AssignmentCardProps> = ({ assignment, submissions, courseID, selfID }) => {
  const history = useHistory()
  const redirectTo = useCallback((submission: Submission) => {
    if (submission.groupID !== 0n) {
      history.push(`/course/${courseID}/group-lab/${submission.AssignmentID.toString()}`)
    } else {
      history.push(`/course/${courseID}/lab/${submission.AssignmentID.toString()}`)
    }
  }, [history, courseID])
  const hasSubmissions = submissions.length > 0
  const redirectToMainAssignment = () => {
    if (hasSubmissions) {
      redirectTo(submissions[0])
    }
  }
  // Add onclick and hover only if there are submissions
  const buttonRole = hasSubmissions ? "button" : ""
  const ariaHidden = hasSubmissions ? "true" : "false"
  const hover = hasSubmissions ? "hover-effect" : ""
  return (
    <div key={assignment.ID.toString()} className={`card mb-4 shadow-sm ${hover}`} onClick={redirectToMainAssignment} role={buttonRole} aria-hidden={ariaHidden}>
      <div className="card-header">
        <div className="d-flex justify-content-between align-items-center">
          <div className="d-flex align-items-center">
            <h5 className="card-title mb-0">{assignment.name}</h5>
            {assignment.isGroupLab && (
              <span className="badge badge-secondary ml-2 p-2">Group</span>
            )}
          </div>
          <div>
            <i className="fa fa-calendar mr-2" /> {getFormattedTime(assignment.deadline, true)}
          </div>
        </div>
      </div>
      <div className="card-body">
        {submissions.map((submission) => (
          <SubmissionRow
            key={submission.ID.toString()}
            submission={submission}
            assignment={assignment}
            courseID={courseID}
            selfID={selfID}
            redirectTo={redirectTo}
          />
        ))}
      </div>
    </div>
  )
}

export default AssignmentCard
