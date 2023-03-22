import React from "react"
import { Submission_Status } from "../../proto/qf/types_pb"
import { Color, isManuallyGraded } from "../Helpers"
import { useActions, useAppState } from "../overmind"
import { ButtonType } from "./admin/Button"
import DynamicButton from "./DynamicButton"

const ManageSubmissionStatus = (): JSX.Element => {
    const actions = useActions()
    const state = useAppState()
    const assignment = state.selectedAssignment

    const [rebuilding, setRebuilding] = React.useState(false)
    const [updating, setUpdating] = React.useState<Submission_Status>(Submission_Status.NONE)


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

    const getButtonType = (status: Submission_Status): ButtonType => {
        if (state.selectedSubmission?.status === status) {
            return ButtonType.BUTTON
        }
        return ButtonType.OUTLINE
    }

    return (
        <div className="row m-auto">
            <DynamicButton
                text="Approve"
                onClick={() => handleSetStatus(Submission_Status.APPROVED)}
                color={Color.GREEN}
                type={getButtonType(Submission_Status.APPROVED)}
                className="col mr-2"
            />
            <DynamicButton
                text="Revision"
                onClick={() => handleSetStatus(Submission_Status.REVISION)}
                color={Color.YELLOW}
                type={getButtonType(Submission_Status.REVISION)}
                className="col mr-2"
            />
            <DynamicButton
                text="Reject"
                onClick={() => handleSetStatus(Submission_Status.REJECTED)}
                color={Color.RED}
                type={getButtonType(Submission_Status.REJECTED)}
                className="col mr-2"
            />
            {assignment && !isManuallyGraded(assignment) && (
                <DynamicButton
                    text={rebuilding ? "Rebuilding..." : "Rebuild"}
                    onClick={handleRebuild}
                    color={Color.BLUE}
                    type={ButtonType.OUTLINE}
                    className="col mr-2"
                />
            )}
        </div>
    )
}

export default ManageSubmissionStatus
