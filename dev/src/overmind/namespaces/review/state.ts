import { derived } from "overmind"
import { Context } from "../.."
import { GradingCriterion, Review, User } from "../../../../proto/ag/ag_pb"

export type ReviewState = {
    /* The index of the selected review */
    selectedReview: number

    /* Contains all reviews for the different courses, indexed by the course id and submission id */
    reviews: {
        [courseID: number]: {
            [submissionID: number]: Review.AsObject[]
        }
    }

    /* The current review */
    // derived from reviews and selectedReview
    currentReview: Review.AsObject | null

    /* The reviewer for the current review */
    // derived from currentReview
    reviewer: User.AsObject | null

    /* Indicates if the current review can be updated */
    canUpdate: boolean

    /* The amount of criteria for the current review */
    criteriaTotal: number

    /* The amount of criteria that have been graded for the current review */
    graded: number

    /* The ID of the assignment selected. Used to determine which assignment to release */
    assignmentID: number

    /* The minimum score submissions must have to be released or approved */
    /* Sent as argument to updateSubmissions */
    minimumScore: number
}

export const state: ReviewState = {
    selectedReview: -1,

    reviews: {},

    currentReview: derived(({ reviews, selectedReview }: ReviewState, rootState: Context["state"]) => {
        if (!(rootState.activeCourse > 0 && rootState.activeSubmission > 0)) {
            return null
        }
        const check = reviews[rootState.activeCourse][rootState.activeSubmission]
        return check ? check[selectedReview] : null
    }),

    reviewer: derived(({ currentReview }: ReviewState, rootState: Context["state"]) => {
        if (!currentReview) {
            return null
        }
        return rootState.users[currentReview.reviewerid]
    }),

    canUpdate: derived(({ currentReview }: ReviewState, rootState: Context["state"]) => {
        return currentReview != null && rootState.activeSubmission > 0 && rootState.activeCourse > 0 && currentReview.id > 0
    }),

    criteriaTotal: derived((state: ReviewState, rootState: Context["state"]) => {
        let total = 0
        if (rootState.currentSubmission, rootState.activeCourse) {
            const assignment = rootState.assignments[rootState.activeCourse]?.find(a => a.id === rootState.currentSubmission?.assignmentid)
            if (assignment) {
                assignment.gradingbenchmarksList.forEach(bm => {
                    bm.criteriaList.forEach(() => {
                        total++
                    })
                })
            }
        }
        return total
    }),

    graded: derived(({ currentReview }: ReviewState) => {
        let total = 0
        currentReview?.gradingbenchmarksList?.forEach(bm => {
            bm.criteriaList.forEach((c) => {
                if (c.grade > GradingCriterion.Grade.NONE) {
                    total++
                }
            })
        })
        return total
    }),

    assignmentID: -1,
    minimumScore: 0,
}
