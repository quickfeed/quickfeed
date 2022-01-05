import { derived, json } from "overmind"
import { Context } from "../.."
import { GradingCriterion, Review, User } from "../../../../proto/ag/ag_pb"

type State = {
    /* The index of the selected review */
    selectedReview: number
    
    /* Contains all reviews for the different courses, indexed by the course id and submission id */
    reviews: { 
        [courseID: number]: {
            [submissionID: number]: Review[]
        }
    }

    /* The current review */
    // derived from reviews and selectedReview
    currentReview: Review | undefined
    
    /* The reviewer for the current review */
    // derived from currentReview
    reviewer: User | undefined
    
    /* Indicates if the current review can be updated */
    canUpdate: boolean

    /* The amount of criteria for the current review */
    criteriaTotal: number
    
    /* The amount of criteria that have been graded for the current review */
    graded: number
}

export const state: State = {
    selectedReview: -1,
    
    reviews: {},
    
    currentReview: derived(({reviews, selectedReview}: State, rootState: Context["state"]) => {
        if (!(rootState.activeCourse > 0 && rootState.activeSubmission > 0)) {
            return undefined
        }
        const check = reviews[rootState.activeCourse][rootState.activeSubmission]
        return check ? check[selectedReview] : undefined
    }),
    
    reviewer: derived(({currentReview}: State, rootState: Context["state"]) => {
        if (!currentReview) {
            return undefined
        }
        return rootState.users[currentReview.getReviewerid()]
    }),
    
    canUpdate: derived(({currentReview}: State, rootState: Context["state"]) => {
        return currentReview != undefined && rootState.activeSubmission > 0 && rootState.activeCourse > 0
    }),

    criteriaTotal: derived((state: State, rootState: Context["state"]) => {
        let total = 0
        if (rootState.currentSubmission, rootState.activeCourse) {
            const assignment = rootState.assignments[rootState.activeCourse]?.find(a => a.getId() === rootState.currentSubmission?.getAssignmentid())
            if (assignment) {
                json(assignment).getGradingbenchmarksList().forEach(bm => {
                    bm.getCriteriaList().forEach(() => {
                        total++
                    })
                })
            }
        }
        return total
    }),

    graded: derived(({currentReview}: State) => {
        let total = 0
        json(currentReview)?.getGradingbenchmarksList()?.forEach(bm => {
            json(bm).getCriteriaList().forEach((c) => {
                console.log(c)
                if (c.getGrade() > GradingCriterion.Grade.NONE) {
                    total++
                }
            })
        })
        return total
    })

}