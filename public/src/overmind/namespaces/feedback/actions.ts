import { create } from "@bufbuild/protobuf"
import { Context } from '../..'
import { CourseRequestSchema } from "../../../../proto/qf/requests_pb"
import { AssignmentFeedback, AssignmentFeedbacks } from '../../../../proto/qf/types_pb'

export const createAssignmentFeedback = async (
    { effects }: Context,
    { feedback }: { courseID: string, feedback: AssignmentFeedback }
): Promise<boolean> => {
    const response = await effects.global.api.client.createAssignmentFeedback(feedback)
    return response.error === null
}

export const getAssignmentFeedback = async (
    { state, effects }: Context,
    { courseID }: { courseID: bigint }
): Promise<AssignmentFeedbacks | null> => {
    const req = create(CourseRequestSchema, {
        courseID,
    })
    const response = await effects.global.api.client.getAssignmentFeedback(req)

    if (response.error) {
        return null
    }

    // Organize feedbacks by assignment within the course
    const byAssignment = new Map<bigint, AssignmentFeedback[]>()
    response.message.feedbacks.forEach(feedback => {
        if (!byAssignment.has(feedback.AssignmentID)) {
            byAssignment.set(feedback.AssignmentID, [])
        }
        byAssignment.get(feedback.AssignmentID)?.push(feedback)
    })

    // Store in state organized by course
    state.feedback.feedback.set(courseID, {
        byAssignment,
        all: response.message.feedbacks
    })

    return response.message
}
