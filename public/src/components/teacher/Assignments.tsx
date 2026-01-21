import React, { useState } from "react"
import { Assignment } from "../../../proto/qf/types_pb"
import {
    isManuallyGraded,
    Color,
    hasBenchmarks,
    getFormattedTime,
} from "../../Helpers"
import { useActions, useAppState } from "../../overmind"
import Button, { ButtonType } from "../admin/Button"
import RubricDisplay from "./RubricDisplay"
import { useCourseID } from "../../hooks/useCourseID"

/** This component displays all assignments for the active course and:
 *  for assignments that are not manually graded, allows teachers to rebuild all submissions.
 *  for manually graded assignments, displays the grading criteria and benchmarks */
const Assignments = () => {
    const courseID = useCourseID()
    const actions = useActions().global
    const state = useAppState()

    const AssignmentElement = ({ assignment }: { assignment: Assignment }) => {
        const [expanded, setExpanded] = useState<boolean>(false)
        const [buttonText, setButtonText] = useState<string>("Rebuild All Tests")
        const [isRebuilding, setIsRebuilding] = useState<boolean>(false)

        const isManual = isManuallyGraded(assignment.reviewers)

        /* rebuild all tests for this assignment */
        const rebuild = async () => {
            if (
                confirm(
                    `Warning! This will rebuild all submissions for ${assignment.name}. This may take several minutes. Are you sure you want to continue?`,
                )
            ) {
                setButtonText("Rebuilding...")
                setIsRebuilding(true)
                const success = await actions.rebuildAllSubmissions({
                    assignmentID: assignment.ID,
                    courseID,
                })
                setIsRebuilding(false)
                if (success) {
                    setButtonText("Rebuild Successful ✓")
                    setTimeout(() => setButtonText("Rebuild All Tests"), 3000)
                } else {
                    setButtonText("Rebuild Failed ✗")
                    setTimeout(() => setButtonText("Rebuild All Tests"), 3000)
                }
            }
        }

        const rubric = hasBenchmarks(assignment) ? (
            <div className="mt-3">
                {assignment.gradingBenchmarks.map((bm) => (
                    <RubricDisplay key={bm.ID.toString()} benchmark={bm} />
                ))}
            </div>
        ) : (
            <div className="alert alert-info mt-3" role="alert">
                <i className="bi bi-info-circle mr-2" />
                No grading benchmarks defined for this assignment.
            </div>
        )

        return (
            <div className="card mb-3 shadow-sm">
                <div
                    className="card-header bg-white"
                    onClick={() => setExpanded(!expanded)}
                    onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault()
                            setExpanded(!expanded)
                        }
                    }}
                    role="button"
                    tabIndex={0}
                    style={{ cursor: "pointer" }}
                >
                    <div className="d-flex justify-content-between align-items-center">
                        <div className="d-flex align-items-center">
                            <h5 className="mb-0 font-weight-bold text-dark">
                                {assignment.name}
                            </h5>
                            <span
                                className={`badge ml-3 ${isManual ? "badge-info" : "badge-secondary"
                                    }`}
                            >
                                {isManual ? "Rubric Grading" : "Auto Graded"}
                            </span>
                        </div>
                        <div className="text-muted">
                            <i className={`bi bi-chevron-${expanded ? "up" : "down"}`} />
                        </div>
                    </div>
                    {assignment.deadline && (
                        <small className="text-muted">
                            Deadline: {getFormattedTime(assignment.deadline, true)}
                        </small>
                    )}
                </div>
                {expanded && (
                    <div className="card-body">
                        {isManual ? (
                            <div>{rubric}</div>
                        ) : (
                            <div className="text-center py-4">
                                <p className="text-muted mb-3">
                                    Rebuild all student submissions for this assignment
                                </p>
                                <Button
                                    text={buttonText}
                                    color={Color.BLUE}
                                    type={ButtonType.BUTTON}
                                    onClick={rebuild}
                                    disabled={isRebuilding}
                                />
                            </div>
                        )}
                    </div>
                )}
            </div>
        )
    }

    return (
        <div className="container-fluid py-4">
            <div className="row">
                <div className="col-12">
                    <h3 className="mb-4">Course Assignments</h3>
                    {state.assignments[courseID.toString()]?.length > 0 ? (
                        state.assignments[courseID.toString()]?.map((assignment) => (
                            <AssignmentElement key={assignment.ID} assignment={assignment} />
                        ))
                    ) : (
                        <div className="alert alert-warning" role="alert">
                            No assignments found for this course.
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}

export default Assignments
