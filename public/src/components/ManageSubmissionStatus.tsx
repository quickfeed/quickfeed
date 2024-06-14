import React from "react"
import { Grade, Submission_Status } from "../../proto/qf/types_pb"
import { Color, hasAllStatus, isManuallyGraded } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { ButtonType } from "./admin/Button"
import DynamicButton from "./DynamicButton"

const ManageSubmissionStatus = (): JSX.Element => {
    const actions = useActions()
    const state = useAppState()
    const assignment = state.selectedAssignment

    const [rebuilding, setRebuilding] = React.useState(false)
    const [updating, setUpdating] = React.useState<Submission_Status>(Submission_Status.NONE)
    const [viewIndividualGrades, setViewIndividualGrades] = React.useState<boolean>(false)

    const handleRebuild = async () => {
        if (rebuilding) { return } // Don't allow multiple rebuilds at once
        setRebuilding(true)
        await actions.rebuildSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission })
        setRebuilding(false)
    }

    const handleSetStatus = async (status: Submission_Status) => {
        if (updating !== Submission_Status.NONE) { return } // Don't allow multiple updates at once
        setUpdating(status)
        await actions.updateSubmission({ owner: state.submissionOwner, submission: state.selectedSubmission, status })
        setUpdating(Submission_Status.NONE)
    }

    const handleSetGrade = async (grade: Grade, status: Submission_Status) => {
        if (updating !== Submission_Status.NONE) { return } // Don't allow multiple updates at once
        setUpdating(status)
        await actions.updateGrade({ grade, status })
        setUpdating(Submission_Status.NONE)
    }

    const getButtonType = (status: Submission_Status): ButtonType => {
        console.log(status)
        const submission = state.selectedSubmission
        const grades = submission?.Grades
        if (!grades) {
            return ButtonType.OUTLINE
        }
        if (hasAllStatus(submission, status)) {
            // Handle in case of multiple grades
            return ButtonType.BUTTON
        }
        //if (state.selectedSubmission?.status === status) {
        return ButtonType.OUTLINE
        //}
        //return ButtonType.OUTLINE
    }

    const getGradeButtonType = (grade: Grade, status: Submission_Status): ButtonType => {
        if (grade.Status === status) {
            return ButtonType.BUTTON
        }
        return ButtonType.OUTLINE
    }

    return (
        <div className="row m-auto">
            <DynamicButton
                text="Approve"
                color={Color.GREEN}
                type={getButtonType(Submission_Status.APPROVED)}
                className="col mr-2"
                onClick={() => handleSetStatus(Submission_Status.APPROVED)}
            />
            <DynamicButton
                text="Revision"
                color={Color.YELLOW}
                type={getButtonType(Submission_Status.REVISION)}
                className="col mr-2"
                onClick={() => handleSetStatus(Submission_Status.REVISION)}
            />
            <DynamicButton
                text="Reject"
                color={Color.RED}
                type={getButtonType(Submission_Status.REJECTED)}
                className="col mr-2"
                onClick={() => handleSetStatus(Submission_Status.REJECTED)}
            />
            {assignment && !isManuallyGraded(assignment) && (
                <DynamicButton
                    text={rebuilding ? "Rebuilding..." : "Rebuild"}
                    color={Color.BLUE}
                    type={ButtonType.OUTLINE}
                    className="col mr-2"
                    onClick={handleRebuild}
                />
            )}
            <DynamicButton
                text={viewIndividualGrades ? "View All Grades" : "View Individual Grades"}
                color={Color.GRAY}
                type={ButtonType.OUTLINE}
                className="col mr-2"
                onClick={async () => setViewIndividualGrades(!viewIndividualGrades)}
            />
            {viewIndividualGrades && state.selectedSubmission?.Grades && state.selectedSubmission.Grades.map((grade) => (
                <div key={grade.UserID.toString()} className="row mt-2">
                    <span className="col">{grade.UserID}</span>
                    <DynamicButton
                        text="Approve"
                        color={Color.GREEN}
                        type={getGradeButtonType(grade, Submission_Status.APPROVED)}
                        className="col mr-2"
                        onClick={() => handleSetGrade(grade, Submission_Status.APPROVED)}
                    />
                    <DynamicButton
                        text="Revision"
                        color={Color.YELLOW}
                        type={getGradeButtonType(grade, Submission_Status.REVISION)}
                        className="col mr-2"
                        onClick={() => handleSetGrade(grade, Submission_Status.REVISION)}
                    />
                    <DynamicButton
                        text="Reject"
                        color={Color.RED}
                        type={getGradeButtonType(grade,Submission_Status.REJECTED)}
                        className="col mr-2"
                        onClick={() => handleSetGrade(grade, Submission_Status.REJECTED)}
                    />
                </div>
            ))}
        </div>
    )
}

export default ManageSubmissionStatus
