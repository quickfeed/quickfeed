import { derived } from "overmind"
import { Context } from "../.."
import { GradingCriterion_Grade, Review, User } from "../../../../proto/qf/types_pb"

export type ReviewState = {
    /* The index of the selected review */
    selectedReview: number

    /* Contains all reviews for the different courses, indexed by the course id and submission id */
    reviews: Map<bigint, Review[]>

    /* The current review */
    // derived from reviews and selectedReview
    currentReview: Review | null

    /* The reviewer for the current review */
    // derived from currentReview
    reviewer: User | null

    /* Indicates if the current review can be updated */
    canUpdate: boolean

    /* The amount of criteria for the current review */
    criteriaTotal: number

    /* The amount of criteria that have been graded for the current review */
    graded: number

    /* The ID of the assignment selected. Used to determine which assignment to release */
    assignmentID: bigint

    /* The minimum score submissions must have to be released or approved */
    /* Sent as argument to updateSubmissions */
    minimumScore: number
}

export const state: ReviewState = {
    selectedReview: -1,

    reviews: new Map(),

    currentReview: derived(({ reviews, selectedReview }: ReviewState, { selectedSubmission, activeCourse }: Context["state"]) => {
        if (!(activeCourse > 0 && selectedSubmission !== null)) {
            return null
        }
        const check = reviews.get(selectedSubmission.ID)
        return check ? check[selectedReview] : null
    }),

    reviewer: derived(({ currentReview }: ReviewState, { users }: Context["state"]) => {
        if (!currentReview) {
            return null
        }
        return users[currentReview.ReviewerID.toString()]
    }),

    canUpdate: derived(({ currentReview }: ReviewState, { activeCourse, selectedSubmission }: Context["state"]) => {
        return currentReview !== null && activeCourse > 0 && currentReview?.ID > 0 && selectedSubmission !== null
    }),

    criteriaTotal: derived((_state: ReviewState, rootState: Context["state"]) => {
        let total = 0
        if (rootState.selectedSubmission && rootState.activeCourse) {
            const assignment = rootState.assignments[rootState.activeCourse.toString()]?.find(a => a.ID === rootState.selectedSubmission?.AssignmentID)
            if (assignment) {
                assignment.gradingBenchmarks.forEach(bm => {
                    bm.criteria.forEach(() => {
                        total++
                    })
                })
            }
        }
        return total
    }),

    graded: derived(({ currentReview }: ReviewState) => {
        let total = 0
        currentReview?.gradingBenchmarks?.forEach(bm => {
            bm.criteria.forEach((c) => {
                if (c.grade > GradingCriterion_Grade.NONE) {
                    total++
                }
            })
        })
        return total
    }),

    assignmentID: BigInt(-1),
    minimumScore: 0,
}
