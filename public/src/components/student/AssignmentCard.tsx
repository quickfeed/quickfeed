import React, { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { getFormattedTime, isValidSubmissionForAssignment } from "../../Helpers"
import { DefaultProgressBar } from '../ProgressBar'
import SubmissionRow from './SubmissionRow'

interface AssignmentCardProps {
  assignment: Assignment
  submissions: Submission[]
  courseID: string
  selfID: bigint
}

const AssignmentCard: React.FC<AssignmentCardProps> = ({ assignment, submissions, courseID, selfID }) => {
  const navigate = useNavigate()
  const redirectTo = useCallback((submission: Submission) => {
    if (submission.groupID !== 0n) {
      navigate(`/course/${courseID}/group-lab/${submission.AssignmentID.toString()}`)
    } else {
      navigate(`/course/${courseID}/lab/${submission.AssignmentID.toString()}`)
    }
  }, [courseID, navigate])
  const validSubmissions = submissions.filter((submission) => isValidSubmissionForAssignment(submission, assignment))
  const hasSubmissions = validSubmissions.length > 0
  const redirectToSubmission = () => {
    if (hasSubmissions) {
      redirectTo(validSubmissions[0])
    }
  }
  // Add onclick and hover only if there are submissions
  const buttonRole = hasSubmissions ? "button" : ""
  const ariaHidden = hasSubmissions ? "true" : "false"
  const hover = hasSubmissions ? "hover-effect" : ""
  return (
    <div key={assignment.ID.toString()} className={`card mb-4 shadow-sm ${hover}`} onClick={redirectToSubmission} role={buttonRole} aria-hidden={ariaHidden}> {/* skipcq: JS-0746, JS-0765 */}
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
        {validSubmissions.map((submission) => (
          <SubmissionRow
            key={submission.ID.toString()}
            submission={submission}
            assignment={assignment}
            courseID={courseID}
            selfID={selfID}
            redirectTo={redirectTo}
          />
        ))}
        {submissions.length === 0 && <DefaultProgressBar scoreLimit={assignment.scoreLimit} isGroupLab={assignment.isGroupLab} />}
      </div>
    </div>
  )
}

export default AssignmentCard
