import { useAppState } from "../overmind"
import { CenteredMessage, KnownMessage } from "./CenteredMessage"
import Lab from "./Lab"
import ManageSubmissionStatus from "./ManageSubmissionStatus"
import Notes from "./Notes"

const LabResult = () => {
    const state = useAppState()
    if (!state.selectedSubmission) {
        return <CenteredMessage message={KnownMessage.TeacherNoSubmission} />
    }
    const assignment = state.selectedAssignment
    if (!assignment) {
        return <CenteredMessage message={KnownMessage.TeacherNoAssignment} />
    }
    return (
        <div className="lab-resize lab-sticky lab-sticky-col">
            <Notes />
            <ManageSubmissionStatus courseID={assignment.CourseID.toString()} reviewers={assignment.reviewers} />
            <div className="reviewLabResult lab-fill mt-2">
                <Lab />
            </div>
        </div>
    )
}

export default LabResult
