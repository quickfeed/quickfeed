import { derived, json } from "overmind"
import { Context } from "../.."
import { GradingCriterion, Review, User } from "../../../../proto/ag/ag_pb"

type State = {
    selectedReview: number
    reviews: { 
        [courseID: number]: {
            [submissionID: number]: Review[]
        }
    }
    currentReview: Review | undefined
    reviewer: User | undefined
    canUpdate: boolean

    // Amount of criteria that are gradable
    criteriaTotal: number
    // Number of graded criteria
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