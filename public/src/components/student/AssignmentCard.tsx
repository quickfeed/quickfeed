import React from 'react'
import { Assignment, Submission } from "../../../proto/qf/types_pb"
import { getFormattedTime } from "../../Helpers"
import SubmissionRow from './SubmissionRow'

interface AssignmentCardProps {
  assignment: Assignment
  submissions: Submission[]
  courseID: string
  selfID: bigint
}

const AssignmentCard: React.FC<AssignmentCardProps> = ({ assignment, submissions, courseID, selfID }) => {
  return (
    <div key={assignment.ID.toString()} className="card mb-4 shadow-sm">
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
          />
        ))}
      </div>
    </div>
  )
}

export default AssignmentCard
