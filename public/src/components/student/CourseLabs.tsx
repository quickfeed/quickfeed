import React from "react"
import { getCourseID } from "../../Helpers"
import { useAppState } from "../../overmind"
import AssignmentCard from "./AssignmentCard"

const CourseLabs = (): JSX.Element => {
  const state = useAppState()
  const courseID = getCourseID().toString()
  const assignments = state.assignments[courseID] || []
  const selfID = state.self.ID

  return (
    <ul className="list-group">
      {assignments.map((assignment) => {
        const submissions = state.submissions.ForAssignment(assignment);
        return (
          <AssignmentCard
            key={assignment.ID.toString()}
            assignment={assignment}
            submissions={submissions}
            courseID={courseID}
            selfID={selfID}
          />
        )
      })}
    </ul>
  )
}

export default CourseLabs
