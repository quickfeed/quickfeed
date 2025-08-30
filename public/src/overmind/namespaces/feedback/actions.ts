import { Context } from '../..'
import { AssignmentFeedback } from '../../../../proto/qf/types_pb'

export const createAssignmentFeedback = async (
    { state, effects }: Context,
    { feedback }: { courseID: string, feedback: AssignmentFeedback }
): Promise<void> => {
    const response = await effects.global.api.client.createAssignmentFeedback(feedback)
    if (response.error) {
        throw new Error('Failed to create assignment feedback')
    }

    // Store the feedback in state
    const feedbackMap = new Map(state.feedback.feedback)
    feedbackMap.set(feedback.AssignmentID, response.message)
    state.feedback.feedback = feedbackMap
}

export const getAssignmentFeedback = async (
    { state, effects }: Context,
    { courseID, assignmentID, userID }: { courseID: string, assignmentID: bigint, userID?: bigint }
): Promise<AssignmentFeedback | null> => {
    const response = await effects.global.api.client.getAssignmentFeedback({
        CourseID: BigInt(courseID),
        AssignmentID: assignmentID,
        UserID: userID || BigInt(0)
    })
    if (response.error) {
        return null
    }

    // Store the feedback in state
    const feedbackMap = new Map(state.feedback.feedback)
    feedbackMap.set(assignmentID, response.message)
    state.feedback.feedback = feedbackMap
    return response.message
}
