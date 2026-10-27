import { create } from "@bufbuild/protobuf"
import { ReviewSchema, SubmissionSchema } from "../../proto/qf/types_pb"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { state as reviewState } from "../overmind/namespaces/review/state"
import { initializeOvermind, mock } from "./TestHelpers"

// The server recomputes the submission's score from its reviews and test
// score, so saving a review must fetch the submission rather than copy the
// review's score into it.
describe("saving a review", () => {
    it("fetches the submission's recomputed score", async () => {
        const review = create(ReviewSchema, { ID: 1n, SubmissionID: 5n, score: 70 })
        const submission = create(SubmissionSchema, { ID: 5n, AssignmentID: 3n, userID: 2n, testScore: 90, score: 54 })
        const api = new ApiClient()
        api.client = {
            ...api.client,
            updateReview: mock("updateReview", async () => ({ message: review, error: null })),
            getSubmission: mock("getSubmission", async () => ({
                message: create(SubmissionSchema, { ...submission, reviewScore: 70, score: 82, reviews: [review] }),
                error: null,
            })),
        }
        const { state, actions } = initializeOvermind({
            activeCourse: 1n,
            selectedSubmission: submission,
            // Spread the initial review state to keep its derived values.
            review: { ...reviewState, reviews: new Map([[submission.ID, [review]]]), selectedReview: 0 },
        }, api)

        expect(await actions.review.updateReview()).toBe(true)
        expect(state.selectedSubmission?.score).toBe(82)
        expect(state.selectedSubmission?.reviewScore).toBe(70)
    })
})
