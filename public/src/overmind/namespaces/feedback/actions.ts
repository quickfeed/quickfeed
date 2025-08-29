import { Context } from '../..'
import { AssignmentFeedback } from '../../../../proto/qf/types_pb'

export const createAssignmentFeedback = async (
    { state, effects }: Context,
    { feedback }: { courseID: string, feedback: AssignmentFeedback }
): Promise<void> => {
    try {
        const response = await effects.global.api.client.createAssignmentFeedback(feedback)
        if (response.error) {
            throw new Error('Failed to create assignment feedback')
        }

        // Store the feedback in state
        const feedbackMap = new Map(state.feedback.feedback)
        feedbackMap.set(feedback.AssignmentID, response.message)
        state.feedback.feedback = feedbackMap
    } catch (error) {
        console.error('Error creating assignment feedback:', error)
        throw error
    }
}

export const getAssignmentFeedback = async (
    { state, effects }: Context,
    { courseID, assignmentID, userID }: { courseID: string, assignmentID: bigint, userID?: bigint }
): Promise<AssignmentFeedback | null> => {
    try {
        const response = await effects.global.api.client.getAssignmentFeedback({
            courseID: BigInt(courseID),
            assignmentID: assignmentID,
            userID: userID || BigInt(0)
        })
        if (response.error) {
            return null
        }

        // Store the feedback in state
        const feedbackMap = new Map(state.feedback.feedback)
        feedbackMap.set(assignmentID, response.message)
        state.feedback.feedback = feedbackMap
        return response.message
    } catch (error) {
        console.error('Error getting assignment feedback:', error)
        return null
    }
}
