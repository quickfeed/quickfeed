import { AssignmentFeedback } from "../../../../proto/qf/types_pb"

// Feedback data structure for a course
export interface CourseFeedbackData {
    byAssignment: Map<bigint, AssignmentFeedback[]>  // Map assignmentID to feedback
    all: AssignmentFeedback[]  // All feedbacks for the course
}

// State for feedback management
export const state = {
    feedback: new Map<bigint, CourseFeedbackData>(), // Map courseID to feedback data
}

export type FeedbackState = typeof state
