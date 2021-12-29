import { derived } from "overmind"
import { Context } from "../.."
import { Review, User } from "../../../../proto/ag/ag_pb"

type State = {
    selectedReview: number
    reviews: { 
        [courseID: number]: {
            [submissionID: number]: Review[]
        }
    }
    currentReview: Review | undefined
    reviewer: User | undefined
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
    })

}