import React from "react"
import { useAppState } from "../../overmind"
import AssignmentCard from "./AssignmentCard"
import { useCourseID } from "../../hooks/useCourseID"

const CourseLabs = () => {
  const state = useAppState()
  const courseID = useCourseID().toString()
  const assignments = state.assignments[courseID] || []
  const selfID = state.self.ID

  return (
    <div className="flex flex-col">
      {assignments.map((assignment) => {
        const submissions = state.submissions.ForAssignment(assignment)
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
    </div>
  )
}

export default CourseLabs
