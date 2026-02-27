import React, { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { getFormattedTime, isValidSubmissionForAssignment } from "../../Helpers"
import { DefaultProgressBar } from '../ProgressBar'
import SubmissionRow from './SubmissionRow'
import Badge from '../Badge'

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
    <div key={assignment.ID.toString()} className={`card mb-4 shadow-sm overflow-hidden ${hover}`} onClick={redirectToSubmission} role={buttonRole} aria-hidden={ariaHidden}> {/* skipcq: JS-0746, JS-0765 */}
      {/* Card Header with Background */}
      <div className="bg-base-200 px-4 py-2 border-b border-base-300">
        <div className="flex flex-row justify-between items-center">
          <div className="flex items-center gap-2">
            <h5 className="card-title mb-0 text-lg font-bold">{assignment.name}</h5>
            {assignment.isGroupLab && (
              <Badge color="yellow" text="Group" type="solid" />
            )}
          </div>
          <div className="flex items-center gap-2 text-sm text-base-content/70">
            <i className="fa fa-calendar" />
            <span>{getFormattedTime(assignment.deadline, true)}</span>
          </div>
        </div>
      </div>

      {/* Card Body */}
      <div className="card-body card-md">
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
