import { derived } from "overmind";
import { GradingCriterion_Grade } from "../../../../proto/qf/types_pb";
export const state = {
    selectedReview: -1,
    reviews: new Map(),
    currentReview: derived(({ reviews, selectedReview }, { selectedSubmission, activeCourse }) => {
        if (!(activeCourse > 0 && selectedSubmission !== null)) {
            return null;
        }
        const check = reviews.get(selectedSubmission.ID);
        return check ? check[selectedReview] : null;
    }),
    reviewer: derived(({ currentReview }, { courseTeachers }) => {
        if (!currentReview) {
            return null;
        }
        return courseTeachers[currentReview.ReviewerID.toString()];
    }),
    canUpdate: derived(({ currentReview }, { activeCourse, selectedSubmission }) => {
        return currentReview !== null && activeCourse > 0 && currentReview?.ID > 0 && selectedSubmission !== null;
    }),
    criteriaTotal: derived((_state, rootState) => {
        let total = 0;
        if (rootState.selectedSubmission && rootState.activeCourse) {
            const assignment = rootState.assignments[rootState.activeCourse.toString()]?.find(a => a.ID === rootState.selectedSubmission?.AssignmentID);
            if (assignment) {
                assignment.gradingBenchmarks.forEach(bm => {
                    bm.criteria.forEach(() => {
                        total++;
                    });
                });
            }
        }
        return total;
    }),
    graded: derived(({ currentReview }) => {
        let total = 0;
        currentReview?.gradingBenchmarks?.forEach(bm => {
            bm.criteria.forEach((c) => {
                if (c.grade > GradingCriterion_Grade.NONE) {
                    total++;
                }
            });
        });
        return total;
    }),
    assignmentID: BigInt(-1),
    minimumScore: 0,
};
