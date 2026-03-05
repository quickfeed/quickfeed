import React, { useEffect, useCallback } from "react"
import { Grade, Submission_Status } from "../../proto/qf/types_pb"
import { Color, hasAllStatus, isManuallyGraded } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { ButtonType } from "./admin/Button"
import DynamicButton from "./DynamicButton"

const ManageSubmissionStatus = ({ courseID, reviewers }: { courseID: string, reviewers: number }) => {
    const actions = useActions().global
    const state = useAppState()

    const [rebuilding, setRebuilding] = React.useState(false)
    const [updating, setUpdating] = React.useState<Submission_Status>(Submission_Status.NONE)
    const [viewIndividualGrades, setViewIndividualGrades] = React.useState<boolean>(false)

    useEffect(() => {
        // reset the view when the selected submission changes
        return () => {
            setViewIndividualGrades(false)
        }
    }, [state.selectedSubmission])

    const handleRebuild = useCallback(async () => {
        if (rebuilding) { return } // Don't allow multiple rebuilds at once
        setRebuilding(true)
        await actions.rebuildSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission })
        setRebuilding(false)
    }, [rebuilding, actions, state.submissionOwner, state.selectedSubmission])

    // handleSetStatusOrGrade updates the grade if it exist and if doesn't it update the submission status
    const handleSetStatusOrGrade = useCallback(async (status: Submission_Status, grade?: Grade) => {
        if (updating !== Submission_Status.NONE) { return } // Don't allow multiple updates at once
        setUpdating(status)
        if (grade) {
            await actions.updateGrade({ grade, status })
        } else {
            await actions.updateSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission, status })
        }
        setUpdating(Submission_Status.NONE)
    }, [updating, actions, state.submissionOwner, state.selectedSubmission])

    const getButtonType = (status: Submission_Status, grade?: Grade) => {
        const submission = state.selectedSubmission
        if (grade?.Status === status || (submission?.Grades && hasAllStatus(submission, status))) {
            return ButtonType.SOLID
        }
        return ButtonType.OUTLINE
    }

    const StatusButtons = ({ grade }: { grade?: Grade }) => {
        const buttonsInfo = [
            { text: "Approve", color: Color.GREEN, status: Submission_Status.APPROVED },
            { text: "Revision", color: Color.YELLOW, status: Submission_Status.REVISION },
            { text: "Reject", color: Color.RED, status: Submission_Status.REJECTED }
        ]

        const dynamicButtons = buttonsInfo.map(({ text, color, status }) => (
            <DynamicButton
                key={text}
                text={text}
                color={color}
                type={getButtonType(status, grade)}
                className={viewIndividualGrades ? "" : "flex-1"}
                onClick={() => handleSetStatusOrGrade(status, grade)}
            />
        ))
        return <>{dynamicButtons}</>
    }

    const getUserName = (userID: bigint): string =>
        state.courseEnrollments[courseID].find(enrollment => enrollment.userID === userID)?.user?.Name ?? ""

    return (
        <div className="space-y-4">
            <div className="flex flex-wrap gap-3">
                {state.selectedSubmission?.Grades && state.selectedSubmission.Grades.length > 1 && (
                    <DynamicButton
                        text={viewIndividualGrades ? "All Grades" : "Individual Grades"}
                        color={Color.GRAY}
                        type={ButtonType.OUTLINE}
                        className="flex-1 min-w-[150px]"
                        onClick={() => Promise.resolve(setViewIndividualGrades(!viewIndividualGrades))}
                    />
                )}
                {!isManuallyGraded(reviewers) && (
                    <DynamicButton
                        text={rebuilding ? "Rebuilding..." : "Rebuild"}
                        color={Color.BLUE}
                        type={ButtonType.OUTLINE}
                        className="flex-1 min-w-[150px]"
                        onClick={handleRebuild}
                    />
                )}
            </div>

            {!viewIndividualGrades && (
                <div className="flex gap-3">
                    <StatusButtons />
                </div>
            )}
            {viewIndividualGrades && (
                <div className="overflow-x-auto">
                    <table className="table table-zebra">
                        <tbody>
                            {state.selectedSubmission?.Grades.map((grade) => (
                                <tr key={grade.UserID.toString()}>
                                    <td className="font-medium">{getUserName(grade.UserID)}</td>
                                    <td>
                                        <div className="flex gap-2 justify-end">
                                            <StatusButtons grade={grade} />
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    )
}

export default ManageSubmissionStatus
