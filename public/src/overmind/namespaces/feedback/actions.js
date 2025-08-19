export const createAssignmentFeedback = async ({ state, effects }, { feedback }) => {
    try {
        const response = await effects.global.api.client.createAssignmentFeedback(feedback);
        if (response.error) {
            throw new Error('Failed to create assignment feedback');
        }
        const feedbackMap = new Map(state.feedback.feedback);
        feedbackMap.set(feedback.AssignmentID, response.message);
        state.feedback.feedback = feedbackMap;
    }
    catch (error) {
        console.error('Error creating assignment feedback:', error);
        throw error;
    }
};
export const getAssignmentFeedback = async ({ state, effects }, { courseID, assignmentID, userID }) => {
    try {
        const response = await effects.global.api.client.getAssignmentFeedback({
            CourseID: BigInt(courseID),
            AssignmentID: assignmentID,
            UserID: userID || BigInt(0)
        });
        if (response.error) {
            return null;
        }
        const feedbackMap = new Map(state.feedback.feedback);
        feedbackMap.set(assignmentID, response.message);
        state.feedback.feedback = feedbackMap;
        return response.message;
    }
    catch (error) {
        console.error('Error getting assignment feedback:', error);
        return null;
    }
};
