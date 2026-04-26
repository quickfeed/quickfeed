import React, { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { getFormattedTime, isGroupSubmission, isValidSubmissionForAssignment } from "../../Helpers"
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
  // Hover effect only applies when the card is clickable (a submission exists to navigate to)
  const hoverClass = hasSubmissions
    ? "cursor-pointer hover:shadow-md hover:ring-2 hover:ring-primary/30 transition-all duration-150"
    : ""
  return (
    <div key={assignment.ID.toString()} className={`card mb-2 shadow-sm overflow-hidden ${hoverClass}`} onClick={redirectToSubmission} role={buttonRole} aria-hidden={ariaHidden}> {/* skipcq: JS-0746, JS-0765 */}
      {/* Card Header */}
      <div className="bg-base-300 px-4 py-1.5 border-b border-base-300">
        <div className="flex flex-row justify-between items-center">
          <div className="flex items-center gap-2">
            <h5 className="card-title mb-0 text-base font-semibold">{assignment.name}</h5>
            {assignment.isGroupLab && (
              <Badge color="yellow" text="Group" type="solid" />
            )}
          </div>
          <div className="flex items-center gap-2 text-xs text-base-content/60">
            <i className="fa fa-calendar" />
            <span>{getFormattedTime(assignment.deadline, true)}</span>
          </div>
        </div>
      </div>

      {/* Card Body */}
      <div className="card-body bg-base-200 p-3">
        {validSubmissions.map((submission) => (
          <div key={submission.ID.toString()} className="hover:bg-base-300 transition-colors duration-150 rounded-md">
            {validSubmissions.length > 1 && (
              <div className="text-xs font-semibold text-base-content/40 uppercase tracking-wider px-2 pt-2 pb-0.5">
                {isGroupSubmission(submission) ? "Group" : "Individual"}
              </div>
            )}
            <SubmissionRow
              submission={submission}
              assignment={assignment}
              selfID={selfID}
              redirectTo={redirectTo}
            />
          </div>
        ))}
        {submissions.length === 0 && <DefaultProgressBar scoreLimit={assignment.scoreLimit} isGroupLab={assignment.isGroupLab} />}
      </div>
    </div>
  )
}

export default AssignmentCard
