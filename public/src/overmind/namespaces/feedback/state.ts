import { AssignmentFeedback } from "../../../../proto/qf/types_pb"

// State for feedback management
export const state = {
    feedback: new Map<bigint, AssignmentFeedback>(), // Map assignmentID to feedback
}

export type FeedbackState = typeof state
